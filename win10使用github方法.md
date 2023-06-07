# github上传项目（windows）

## 1、[下载git-windows版本](https://git-scm.com/downloads)，并安装

## 2、在github中创建一个repository

## 3、确认**user.name**与**user.email**与自己github账号名称及邮箱是否一致

<img src="images\image-20230508104127292.png" alt="image-20230508104127292" style="zoom:33%;" />

若不一致使用下面命令修改**user.name**与**user.email**

```bash
git config --global user.name xxx
git config --global user.email http://xx.com
# 再次确定user.name与user.email
git config --list
```

## 4、生成ssh-key 

```
ssh-keygen -t rsa -C http://xxx.com
```

一路回车，或者输入y，最终会在如下目录中生成id_rsa等文件

![image-20230508104545207](images\image-20230508104545207.png)

## 5、打开github设置界面

<img src="images\image-20230508104642129.png" alt="image-20230508104642129" style="zoom:33%;" />

## 6、新增ssh key

点击下图中环框按钮

![image-20230508104849329](images\image-20230508104849329.png)



![image-20230508105054829](images\image-20230508105054829.png)

## 7、本地创建一个git项目目录

本例目录为： E:\study\Document\k8s\云原生架构师课程\CN-operator

进入目录后右击打开git bash后，克隆项目到本地

```
zy835@YONX MINGW64 /e/study/Document/k8s/云原生架构师课程/CN-operator (main)
$ git clone git@github.com:SiRain/CN-operator.git
Cloning into 'CN-operator'...
warning: You appear to have cloned an empty repository.
```

## 8、上传本地文档到github

将项目代码所需的代码及文件放入clone的项目文件夹下，在Git Bash界面中依次输入

```bash
cd XXX # XXX为项目名称，cd进入clone的项目文件夹
git add .
git commit -m "update files"
git push
```
