
Here is an example how you can use `Arena` for the machine learning training. It will download the source code from git url, and use Tensorboard to visualize the Tensorflow computation graph and plot quantitative metrics.

1. the first step is to check the available resources

```
arena top node
NAME                                IPADDRESS      ROLE    GPU(Total)  GPU(Allocated)
i-j6c68vrtpvj708d9x6j0  192.168.1.116  master  0           0
i-j6c8ef8d9sqhsy950x7x  192.168.1.119  worker  1           0
i-j6c8ef8d9sqhsy950x7y  192.168.1.120  worker  1           0
i-j6c8ef8d9sqhsy950x7z  192.168.1.118  worker  1           0
i-j6ccue91mx9n2qav7qsm  192.168.1.115  master  0           0
i-j6ce09gzdig6cfcy1lwr  192.168.1.117  master  0           0
-----------------------------------------------------------------------------------------
Allocated/Total GPUs In Cluster:
0/3 (0%)
```

There are 3 available nodes with GPU for running training jobs.


2\. Now we can submit a training job with `arena cli`, it will download the source code from github

```
# arena submit tf \
             --name=tf-tensorboard \
             --gpus=1 \
             --image=tensorflow/tensorflow:1.5.0-devel-gpu \
             --env=TEST_TMPDIR=code/tensorflow-sample-code/ \
             --syncMode=git \
             --syncSource=https://github.com/cheyang/tensorflow-sample-code.git \
             --tensorboard \
             --logdir=/training_logs \
             "python code/tensorflow-sample-code/tfjob/docker/mnist/main.py --max_steps 5000"
configmap/tf-tensorboard-tfjob created
configmap/tf-tensorboard-tfjob labeled
service/tf-tensorboard-tensorboard created
deployment.extensions/tf-tensorboard-tensorboard created
tfjob.kubeflow.org/tf-tensorboard created
INFO[0001] The Job tf-tensorboard has been submitted successfully
INFO[0001] You can run `arena get tf-tensorboard --type tfjob` to check the job status
```

> the source code will be downloaded and extracted to the directory `code/` of the working directory. The default working directory is `/root`, you can also specify by using `--workingDir`.

> `logdir` indicates where the tensorboard reads the event logs of TensorFlow

3\. List all the jobs

```
# arena list
NAME            STATUS     TRAINER  AGE  NODE
tf-tensorboard  RUNNING    TFJOB    0s   192.168.1.119
```

4\. Check the resource usage of the job

```
# arena top job
NAME            STATUS     TRAINER  AGE  NODE           GPU(Requests)  GPU(Allocated)
tf-tensorboard  RUNNING    TFJOB    26s  192.168.1.119  1              1


Total Allocated GPUs of Training Job:
0

Total Requested GPUs of Training Job:
1
```



5\. Check the resource usage of the cluster


```
# arena top node
NAME                    IPADDRESS      ROLE    GPU(Total)  GPU(Allocated)
i-j6c68vrtpvj708d9x6j0  192.168.1.116  master  0           0
i-j6c8ef8d9sqhsy950x7x  192.168.1.119  worker  1           1
i-j6c8ef8d9sqhsy950x7y  192.168.1.120  worker  1           0
i-j6c8ef8d9sqhsy950x7z  192.168.1.118  worker  1           0
i-j6ccue91mx9n2qav7qsm  192.168.1.115  master  0           0
i-j6ce09gzdig6cfcy1lwr  192.168.1.117  master  0           0
-----------------------------------------------------------------------------------------
Allocated/Total GPUs In Cluster:
1/3 (33%)
```


6\. Get the details of the specific job

```
# arena get tf-tensorboard
NAME            STATUS   TRAINER  AGE  INSTANCE                               NODE
tf-tensorboard  RUNNING  tfjob    15s  tf-tensorboard-tfjob-586fcf4d6f-vtlxv  192.168.1.119
tf-tensorboard  RUNNING  tfjob    15s  tf-tensorboard-tfjob-worker-0          192.168.1.119

Your tensorboard will be available on:
192.168.1.117:30670
```

> Notice: you can access the tensorboard by using `192.168.1.117:30670`. You can consider `sshuttle` if you can't access the tensorboard directly from your laptop. For example: `sshuttle -r root@47.89.59.51 192.168.0.0/16`


![](2-tensorboard.jpg)

Congratulations! You've run the training job with `arena` successfully, and you can also check the tensorboard easily.