#!/bin/bash
set -e

# Usage:
#   rsync_parallel.sh [--parallel=N] [rsync args...]
# 
# Options:
#   --parallel=N	Use N parallel processes for transfer. Defaults to 10.
#
# Notes:
#   * Requires GNU Parallel
#   * Use with ssh-keys. Lots of password prompts will get very annoying.
#   * Does an itemize-changes first, then chunks the resulting file list and launches N parallel
#     rsyncs to transfer a chunk each.
#   * be a little careful with the options you pass through to rsync. Normal ones will work, you 
#     might want to test weird options upfront.
#
# ./rsync_parallel.sh -av --ignore-existing —-progress /Volumes/HardDrive/go/src/github.com/lucmichalski/cars-dataset/public ubuntu@35.179.44.166:/home/ubuntu/cars-dataset/
if [[ "$1" == --parallel=* ]]; then
	PARALLEL="${1##*=}"
	shift
else
	PARALLEL=10
fi
echo "Using up to $PARALLEL processes for transfer..."

TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

echo "Figuring out file list..."
# sorted by size (descending)
rsync $@ --out-format="%l %n" --no-v --dry-run | sort -n -r > $TMPDIR/files.all

# check for nothing-to-do
TOTAL_FILES=$(cat $TMPDIR/files.all | wc -l)
if [ "$TOTAL_FILES" -eq "0" ]; then
	echo "Nothing to transfer :)"
	exit 0
fi

function array_min {
  ARR=("$@")

  # Default index for min value
  min_i=0

  # Default min value
  min_v=${ARR[$min_i]}

  for i in "${!ARR[@]}"; do
    v="${ARR[$i]}"

    (( v < min_v )) && min_v=$v && min_i=$i
  done

  echo "${min_i}"
}

echo "Calculating chunks..."
# declare chunk-size array
for ((I = 0 ; I < PARALLEL ; I++ )); do
	CHUNKS["$I"]=0 
done

# add each file to the emptiest chunk, so they're as balanced by size as possible
PROGRESS=0
SECONDS=0
while read FSIZE FPATH; do
  PROGRESS=$((PROGRESS+1))

  # Original Implementation
  #MIN=($(array_min_old ${CHUNKS[@]})); MIN_I=${MIN[0]}
  # Nathan's implementation
  MIN_I=$(array_min ${CHUNKS[@]})

  CHUNKS[${MIN_I}]=$((${CHUNKS[${MIN_I}]} + ${FSIZE}))
  echo "${FPATH}" >> "${TMPDIR}/chunk.${MIN_I}"

  if ! ((PROGRESS % 5000)); then
    >&2 echo "${SECONDS}s: ${PROGRESS} of ${TOTAL_FILES}"
  fi
done < "${TMPDIR}/files.all"
echo "${SECONDS}s"

cp -R $TMPDIR/* ./files

find "$TMPDIR" -type f -name "chunk.*" -exec cat {} \;

echo "Starting transfers..."
find "$TMPDIR" -type f -name "chunk.*" | parallel -j $PARALLEL -t --verbose --progress rsync -e "ssh -i /Users/lucmichalski/Downloads/ounsi.pem" --files-from={} $@
