build:
	@docker build -t labelme2yolo .

run:
	@docker run -ti -v `pwd`/train:/opt/data labelme2yolo --data-dir=/opt/data
