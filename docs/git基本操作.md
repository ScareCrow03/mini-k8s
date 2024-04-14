## git基本操作

### git不同区域

远程仓库

本地仓库

暂存区

工作区

![img](https://pic2.zhimg.com/80/v2-3bc9d5f2c49a713c776e69676d7d56c5_1440w.webp)

fork的时候，原仓库为upstream，fork出来的仓库为origin

自己建的远程仓库叫做origin

在本地看到的目录是工作区

### git指令

#### 工作区到本地仓库

这是为了把新修改的文件给保存到仓库中

为git add和git commit

git add .将工作区中所有的文件加入暂存区中，可以设定把哪些文件放入暂存区中

git commit -m "消息" 把暂存区中的文件放入本地仓库中进行保存

#### 本地仓库到工作区

一种是恢复文件，一种是切换分支

git checkout -- [filename]表示从仓库中恢复某个文件

git checkout [branchname] 表示转到某一个分支

#### 本地仓库到远程仓库

git push <远程主机名> <本地分支名>:<远程分支名>

远程主机名默认为origin，当本地分支和远程分支名相同时，可以省略一个，如果已经存在本地分支和远程分支的追踪关系，那么本地分支和远程分支都可以不用写，将现在所处的分支进行提交

使用 -u 参数可以让本地分支和远程分支建立联系

#### 远程仓库到本地仓库

git fetch <远程仓库名字> 将远程仓库的所有信息都放入本地仓库中

git fetch <远程仓库名> <分支名> 将远程仓库的特定分支放入本地仓库中

#### 远程仓库与本地分支合并

git pull <远程仓库名> <远程分支名>:<本地分支名> 把远程仓库某个分支的更新取回，并与本地某个分支进行合并

如果就是本分支，那么可以省略本地分支名

**git fetch不会将本地代码进行合并，只会将仓库里记录remote commitID的修改，而git pull会直接修改本地的文件**

#### 分支创建

git branch [branchname] 创建一个新分支

git checkout -b [branchname] 创建一个新分支，并且跳到当前分支中

git checkout -b [branchname] [remotebranchname] 创建于给新分支，该分支与远程仓库中某个分支相同

git branch --set-upstream [branchname] [remote branchname]将本地分支和远程某分支进行关联（可以用来创建新的远程分支)

git push origin -delete [remote branchname]删除远程分支

#### 与远程仓库连接

git clone url