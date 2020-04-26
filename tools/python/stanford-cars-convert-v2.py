"""
CONVERT .MAT file to .CSV for the stanford-cars-dataset.

http://ai.stanford.edu/~jkrause/cars/car_dataset.html


This script extract the:

  TRAIN/VALIDATION/TEST
  label/class
  filename
  bbox_x1: Min x-value of the bounding box, in pixels
  bbox_x2: Max x-value of the bounding box, in pixels
  bbox_y1: Min y-value of the bounding box, in pixels
  bbox_y2: Max y-value of the bounding box, in pixels
 
from the cars_train_annos.mat file in  the stanford-cars-dataset.zip

# Output is a .csv like:

TRAIN,data/train/cars_train/02443.jpg,74,62,617,411,HUMMER H3T Crew Cab 2010
TRAIN,data/train/cars_train/02444.jpg,70,60,737,541,Ford F-150 Regular Cab 2012
TRAIN,data/train/cars_train/02445.jpg,30,99,743,427,Buick Rainier SUV 2007
TRAIN,data/train/cars_train/02446.jpg,60,126,839,578,Lamborghini Diablo Coupe 2001
TRAIN,data/train/cars_train/02447.jpg,64,58,341,177,Ram C/V Cargo Van Minivan 2012
TRAIN,data/train/cars_train/02448.jpg,3,226,1017,621,Dodge Dakota Club Cab 2007
TRAIN,data/train/cars_train/02449.jpg,61,150,595,378,Honda Odyssey Minivan 2012
TRAIN,data/train/cars_train/02450.jpg,51,99,595,382,Buick Verano Sedan 2012
.
.
.

"""

# encoding:utf8

from scipy.io import loadmat
import pandas as pd
import numpy as np
import os
import struct

mat_train = loadmat('../shared/dataset/stanford-cars/devkit/cars_train_annos.mat')
mat_test = loadmat('../shared/dataset/stanford-cars/devkit/cars_test_annos.mat')
meta = loadmat('../shared/dataset/stanford-cars/devkit/cars_meta.mat')

labels = list()

class UnknownImageFormat(Exception):
    pass

def get_image_size(file_path):
    """
    Return (width, height) for a given img file content - no external
    dependencies except the os and struct modules from core
    """
    size = os.path.getsize(file_path)

    with open(file_path) as input:
        height = -1
        width = -1
        data = input.read(25)

        if (size >= 10) and data[:6] in ('GIF87a', 'GIF89a'):
            # GIFs
            w, h = struct.unpack("<HH", data[6:10])
            width = int(w)
            height = int(h)
        elif ((size >= 24) and data.startswith('\211PNG\r\n\032\n')
              and (data[12:16] == 'IHDR')):
            # PNGs
            w, h = struct.unpack(">LL", data[16:24])
            width = int(w)
            height = int(h)
        elif (size >= 16) and data.startswith('\211PNG\r\n\032\n'):
            # older PNGs?
            w, h = struct.unpack(">LL", data[8:16])
            width = int(w)
            height = int(h)
        elif (size >= 2) and data.startswith('\377\330'):
            # JPEG
            msg = " raised while trying to decode as JPEG."
            input.seek(0)
            input.read(2)
            b = input.read(1)
            try:
                while (b and ord(b) != 0xDA):
                    while (ord(b) != 0xFF): b = input.read(1)
                    while (ord(b) == 0xFF): b = input.read(1)
                    if (ord(b) >= 0xC0 and ord(b) <= 0xC3):
                        input.read(3)
                        h, w = struct.unpack(">HH", input.read(4))
                        break
                    else:
                        input.read(int(struct.unpack(">H", input.read(2))[0])-2)
                    b = input.read(1)
                width = int(w)
                height = int(h)
            except struct.error:
                raise UnknownImageFormat("StructError" + msg)
            except ValueError:
                raise UnknownImageFormat("ValueError" + msg)
            except Exception as e:
                raise UnknownImageFormat(e.__class__.__name__ + msg)
        elif (file_path.endswith('.ico')):
          #see http://en.wikipedia.org/wiki/ICO_(file_format)
          input.seek(0)
          reserved = input.read(2)
          if 0 != struct.unpack("<H", reserved )[0]:
            raise UnknownImageFormat("Corrupt ICON File")
          format = input.read(2)
          assert 1 == struct.unpack("<H", format)[0]
          num = input.read(2)
          num = struct.unpack("<H", num)[0]
          if num > 1:
            import warnings
            warnings.warn("ICO File contains more than one image")
          #http://msdn.microsoft.com/en-us/library/ms997538.aspx
          w = input.read(1) 
          h = input.read(1) 
          width = ord(w)
          height = ord(h)
        else:
          raise UnknownImageFormat(
                "Sorry, don't know how to get information from %s." % file_path
            )

    return width, height

def convert(size, box):
    dw = 1./size[0]
    dh = 1./size[1]
    x = (box[0] + box[1])/2.0
    y = (box[2] + box[3])/2.0
    w = box[1] - box[0]
    h = box[3] - box[2]
    x = x*dw
    w = w*dw
    y = y*dh
    h = h*dh
    return (x,y,w,h)

for l in meta['class_names'][0]:
    labels.append(l[0])
    
train = list()
for example in mat_train['annotations'][0]:
    label = labels[example[-2][0][0]-1]
    image = example[-1][0]
    bbox_x1 = example[0][0][0]
    bbox_x2 = example[1][0][0]
    bbox_y1 = example[2][0][0]
    bbox_y2 = example[3][0][0]
    train.append((image,bbox_x1, bbox_x2, bbox_y1, bbox_y2, label))
    
test = list()
for example in mat_test['annotations'][0]:
    image = example[-1][0]
    test.append(image)

validation_size = int(len(train) * 0.10)
test_size = int(len(train) * 0.20)

validation = train[:validation_size].copy()
np.random.shuffle(validation)
train = train[validation_size:]

test = train[:test_size].copy()
np.random.shuffle(test)
train = train[test_size:]

# Google drive mount example or local path

train_path = '../shared/dataset/stanford-cars/train/cars_train/'
test_path = '../shared/dataset/stanford-cars/test/cars_test/'

f.write('TRAIN,image_path,width,height,bbox_x1,bbox_x2,bbox_y1,bbox_y2,lab,\n')
with open('../shared/dataset/stanford-cars/cars_data.csv', 'w+') as f:
    w, h = get_image_size(img)
    b = (bbox_x1, bbox_x2, bbox_y1, bbox_y2)
    bb = convert((w,h), b)
    print(bb)
    [f.write('TRAIN,%s%s,%s,%s,%s,%s,%s,%s\n' %(train_path, img, w, h, bbox_x1, bbox_x2, bbox_y1, bbox_y2, lab)) for img, bbox_x1, bbox_x2, bbox_y1, bbox_y2, lab in train, " ".join([str(a) for a in bb])]
    [f.write('VALIDATION,%s%s,%s,%s,%s,%s,%s,%s\n' %(train_path, img, w, h, bbox_x1, bbox_x2, bbox_y1, bbox_y2, lab)) for img, bbox_x1, bbox_x2, bbox_y1, bbox_y2, lab in validation, " ".join([str(a) for a in bb])]
    # [f.write('TEST,%s%s\n' %(test_path,img)) for img,_,_,_,_,_,_,_ in test]

# encoding:utf8