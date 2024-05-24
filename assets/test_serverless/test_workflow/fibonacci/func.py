# 这个函数接受三个参数x、y和i，并返回一个字典，其中x的值为y，y的值为x+y，i的值为i+1。这个函数可以用于计算斐波那契数列的下一个值，即传入、传出的三元组可以用于描述斐波那契数列的状态，满足x=fib(i-1)、y=fib(i)、i是当前已经计算好的下标
def handle(param):
    x = param["x"]
    y = param["y"]
    i = param["i"]
    ret = dict()
    ret["x"] = y
    ret["y"] = x + y
    ret["i"] = i + 1
    return ret
