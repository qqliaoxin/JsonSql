<h1 align="center" style="text-align:center;">
  JSONSQL
</h1>
 
<p align="center">零代码、热更新Table、自动化 ORM 库<br />🚀 后端接口零代码，前端(客户端) 定制返回 JSON 的数据和结构</p>

<p align="center" >
    </p>
    <h1 align="center" style="text-align:center;">
    <font size = 8>
    {sql@}
    </font>
    </h1>
    </p>
</p>


# 极简的Json 转换SQL 查询风格
最简的demo, 需要在config.json -> Table中添加配置映射关系。例如：
```
{
    "Table":{
        "Test":{
            "TableName":"mf_test",
        }

    },
    "Debug": true
}
```
postman 请求连接 http://localhost:8080/get ，请求方式 Post Body 请求json内容，如下： 
```
{
    "Test":{
        "@w":{
            "id": 1
        }
    }
}
```
生成的Mysql SQL语句:
```
SELECT * FROM mf_test AS test WHERE  id = ?
```
返回的结果:
```
{"Test":[{"date":"2021-08-02 13:45:16","detail":"test","id":1}],"code":200}
```
# JsonSql 关键核心转义对象

|  关键字符号   | 代表的sql查询转义  |
|  ----  | ----  |
| xxx@  | Json一级对象转义为查询的映射对象, 带@是引用对象， 带|为关联表引用 |
| @column 简写:@c  | Sql 语句SELECT 要返回的字段,字段可以挟带sql语句。 |
| @where 简写:@w  | WHERE 语句后的条件生成对象 |
| @group 简写:@g  | GROUP BY  |
| @order 简写:@o  | ORDER BY |
| @limit 简写:@l  | LIMIT |
| @offset 简写:@os  | OFFSET |

# JsonSql 相关符号使用含义

|  @where value 符号   | 转义  |
|  ----  | ----  |
| @  | 引用符号，可以是引用或者是构建[table]查询生成对象,同时也会触发对value值的含义转换，如: User/id -> user.id |
| /  | 转义，@column 中转义为 AS ,@where 中转义为 . |
| &  | 转义 AND |
| |  | OR |
| %  | 应用与字段后缀， 转义 LIKE '%%' |
| >=,>,<,<=  | 与sql语句 中的含义一样 |
| []  | 转义 IN  ,允许写法  [1,2] ,["user1",user2] , "sql@" |


# idea 还可以像sql 语句的逻辑，构造关联查询语句。
子查询
```
{
    "sql@":{
        "Test":{
            "@column": "nick_name",
            "@where":{
                "id": 1
            }
        }
    },
    "User|Test":{
        "@column":"id,detail AS Detail,sql@/NickName",
        "@where":{
            "User/id@": "Test/id",
            "&User/sex@": "boy",
            "|kks": 100, 
            "&detail%": "搜索内容",
            "&id": 12345
        },
        "@limit": 10,
        "@offset" : 2,
        "@order":"date-"
    }
}
```
生成sql语句:
```
SELECT id,detail AS Detail,(SELECT nick_name FROM mf_test AS test WHERE  id = ?) AS NickName FROM apijson_user AS user, mf_test AS test WHERE  user.id = test.id AND  user.sex = ? OR  kks = ? AND  detail LIKE '%?%' AND  id = ? ORDER BY date DESC LIMIT 10 OFFSET  2
```
多种复杂组合， bsql@ 复合引用 asql@, 命名规则已 ASCII 为优先级排序生成 sql语句。
```
{
    "asql@":{
        "Test":{
            "@column": "nick_name",
            "@where":{
                "id": 1
            }
        }
    },
    "bsql@":{
        "Test":{
            "@column": "nick_name_2",
            "@where":{
                "id@": "asql@",
                "Test/name@":"klkk"
            }
        }
    },
    "User|Test":{
        "@column":"id,detail AS Detail,bsql@/NickName",
        "@where":{
            "User/id@": "Test/id",
            "User/sex@": "boy",
            "|kks": 100, 
            "&detail%": "搜索内容",
            "&id": 12345
        },
        "@limit": 10,
        "@offset" : 2,
        "@order":"date-"
    }
}
```
生成的sql语句：
```
SELECT id,detail AS Detail,(SELECT nick_name_2 FROM mf_test AS test WHERE  id IN (SELECT nick_name FROM mf_test AS test WHERE  id = ?) AND  test.name = ?) AS NickName FROM apijson_user AS user, mf_test AS test WHERE  user.id = test.id AND  user.sex = ? OR  kks = ? AND  detail LIKE '%?%' AND  id = ? ORDER BY date DESC LIMIT 10 OFFSET  2
```

还可以做为关联表
```
{
    "asql@":{
        "Test":{
            "@column": "nick_name",
            "@where":{
                "id": 1
            }
        }
    },
    "bsql@":{
        "Test":{
            "@column": "nick_name_2",
            "@where":{
                "id@": "asql@",
                "Test/name@":"klkk"
            }
        }
    },
    "User|asql@":{
        "@column":"id,detail AS Detail,NickName",
        "@where":{
            "User/id@": "asql/id",
            "User/sex@": "boy",
            "|kks": 100, 
            "&detail%": "搜索内容",
            "&id": 12345
        },
        "@limit": 10,
        "@offset" : 2,
        "@order":"date+"
    }
}
```
生成的sql语句：
```
SELECT id,detail AS Detail,NickName FROM apijson_user AS user, (SELECT nick_name FROM mf_test AS test WHERE  id = ?) AS asql WHERE  user.sex = ? AND  user.id = asql.id OR  kks = ? AND  detail LIKE '%?%' AND  id = ? ORDER BY date ASC LIMIT 10 OFFSET  2
```
[] In条件查询的特殊操作:
```
{
    "sql@":{
        "User":{
            "@c": "id",
            "@w":{
                "userid":1
            }
        }
    },
    "Test":{
        "@column": "*",
        "@where":{
            "detail[]": "sql@",
            "id[]":[1,2,3,4]
        }
    }
}
```
生成的sql语句：
```
SELECT * FROM mf_test AS test WHERE  detail IN (SELECT id FROM apijson_user AS user WHERE  userid = ?) AND  id IN (1,2,3,4)
```
# Inster 插入 请求连接 http://localhost:8080/set
```
{
    "Test":{
        "@column": "id,name,pwd",
        "@values": [121,"ssdfaf","paw1233456"]
    }
}
```
生成的sql语句：
```
INSERT INTO  mf_test (id,name,pwd) VALUES(?,?,?)
```
OR
```
{
    "sql@":{
        "User":{
            "@c": "id",
            "@w":{
                "userid":1
            }
        }
    },
    "Test":{
        "@column": "userId,detail",
        "User":{
            "@column": "id,desc",
            "@where":{
             "detail[]": "sql@",
            "id[]":[1,2,3,4]
            }
        }
    }
}
```
生成的sql语句：
```
INSERT INTO  mf_test (userId,detail) SELECT id,desc FROM apijson_user AS user WHERE  detail IN (SELECT id FROM apijson_user AS user WHERE  userid = ?) AND  id IN (1,2,3,4)
```
# Update 更新 请求连接 http://localhost:8080/up
```
{
    "Test":{
        "@set": {
            "userId": 411,
            "detail": "kkkkkk"            
        },
        "@where": {
            "id": 1
        }
    }
}
```
生成的sql语句：
```
UPDATE  mf_test SET userId = ?,detail = ? WHERE  id = ?
```
# Delete 删除 请求连接 http://localhost:8080/del
```
{
    "Test":{
        "@where": {
            "id": 1
        }
    }
}
```
生成的sql语句：
```
DELETE FROM  mf_test WHERE  id = ?
```
## 加入社区

扫码加入即刻交流与反馈：

<img alt="Join the chat at dingtalk" src="./image/jsonsql@.jpg" width="200">