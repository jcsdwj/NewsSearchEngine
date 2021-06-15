#coding:utf-8
import pandas as pd
import numpy as np
from sklearn.metrics import pairwise_distances
import json

#def cknm(filepath,dt_matrix,k):
def cknm(filepath,k):
    # 传入的分别为矩阵dt_matrix和整数k
    # 读取文件 根据文件构建二位数组
    k = int(k) #这里的k为字符串类型

    #dt_matrix=strToarray(dt_matrix) #传过来的是字符串
    dt_matrix=filetoarray(filepath)
    k_nrearest=[]
    ndt_matrix = pd.DataFrame(dt_matrix) #使用二维数组构建矩阵
    ndt_matrix.index = ndt_matrix[0] #这里的矩阵要为n*n

    tmp=np.array(1-pairwise_distances(ndt_matrix[ndt_matrix.columns[1:]],metric='cosine'))
    #similarity_matrix = pd.DataFrame(tmp,index=ndt_matrix.index.tolist(),columns=ndt_matrix.index.tolist())
    similarity_matrix = pd.DataFrame(tmp,index=ndt_matrix.index.tolist(),columns=ndt_matrix.index.tolist())
    for i in similarity_matrix.index:
        tmp = [int(i), []]
        j = 0
        while j < k:
            max_col = similarity_matrix.loc[i].idxmax(axis = 1) #这里为什么不是整数
            similarity_matrix.loc[i][max_col] = -1
            if max_col != i:
                tmp[1].append(int(max_col))  # max column name
                j += 1
        k_nrearest.append(tmp)
    k_dict = {}
    for _ in k_nrearest:#变成字典
        k_dict[str(_[0])] = _[1]
    return json.dumps(k_dict) #将字典转为json格式

def strToarray(Str): #字符串转数组
    l=[]
    m=[]
    s=""
    for x in Str:
        if x == '\n':
            l.append(m)
            m=[]
        elif x!='\t' and x != '\n':
            s+=x
            #m.append(int(x))
        elif x=='\t':
            m.append(int(s))
            s = ""
    return l

def filetoarray(filepath):
    file = open(filepath)
    l=[]
    m=[]
    s=""
    while 1:
        line = file.readline() #读一行
        if not line:
            break
        for x in line:
            if x == '\n':
                l.append(m)
                m=[]
            elif x!='\t' and x != '\n':
                s+=x
            elif x=='\t':
                m.append(int(s))
                s = ""
    return l