# LabelmeToYolo
YOLO-Mark is not a good tool to get labels of the training sets for object detection,So we use labelme to get the labels, then transform them to the format of YOLO.

this program only need one parameter --data_dir. 

this is your directory which contains the images and the json files.the json files are generated after you use labelme to annotate images.

the json files contain the information of the polygon, the aim of this program is contvert the json files to the information of bounding boxes which are needed in YOLO training.

the command is like this 
```bash
python labelme2yolo.py --data_dir D:\your_derictory\"
```

after running this program you will see a directrory at your data directory's parent derictory named 'yolo_need' and a text file named 'yolo_train.txt' both of them are needed for YOLO training.
