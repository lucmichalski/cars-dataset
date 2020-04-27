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

# yolo annotation
0 0.5 0.5 0.10 0.25
represents an object of class 0, centered in the middle of the image, whose width is 10% of the image, and whose height is 25% of the image.

"""

# encoding:utf8

from scipy.io import loadmat
import pandas as pd
import numpy as np
import get_image_size

mat_train = loadmat('../../shared/datasets/stanford-cars/devkit/cars_train_annos.mat')
mat_test = loadmat('../../shared/datasets/stanford-cars/devkit/cars_test_annos.mat')
meta = loadmat('../../shared/datasets/stanford-cars/devkit/cars_meta.mat')
labels = list()

def image_size(file_path):
  try:
    width, height = get_image_size.get_image_size(file_path)
  except get_image_size.UnknownImageFormat:
    width, height = -1, -1

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

def yolo_annotation(img, bbox_x1, bbox_x2, bbox_y1, bbox_y2):
    w, h = image_size(img)
    b = (bbox_x1, bbox_y1, bbox_x2, bbox_y2)
    # b = (bbox_x1, bbox_x2, bbox_y1, bbox_y2)
    bb = convert((w,h), b)
    print(bb)
    return bb

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

train_path = '../../shared/datasets/stanford-cars/cars_train/'
test_path = '../../shared/datasets/stanford-cars/cars_test/'

# f.write('TRAIN,image_path,width,height,bbox_x1,bbox_x2,bbox_y1,bbox_y2,lab,\n')
with open('../../shared/datasets/stanford-cars/yolo_cars_data.csv', 'w+') as f:
    [f.write('TRAIN;%s%s;%s;%s;%s;%s;%s;%s;%s\n' %(train_path, img, image_size(train_path+img), bbox_x1, bbox_x2, bbox_y1, bbox_y2, lab, yolo_annotation(train_path+img, bbox_x1, bbox_x2, bbox_y1, bbox_y2))) for img, bbox_x1, bbox_x2, bbox_y1, bbox_y2, lab in train]
    [f.write('VALIDATION;%s%s;%s,%s;%s;%s;%s;%s;%s\n' %(train_path, img, image_size(test_path+img), bbox_x1, bbox_x2, bbox_y1, bbox_y2, lab, yolo_annotation(test_path+img, bbox_x1, bbox_x2, bbox_y1, bbox_y2))) for img, bbox_x1, bbox_x2, bbox_y1, bbox_y2, lab in validation]
    # [f.write('TEST,%s%s\n' %(test_path,img)) for img,_,_,_,_,_,_,_ in test]

# encoding:utf8
