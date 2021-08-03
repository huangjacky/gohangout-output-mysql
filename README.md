# gohangout-output-mysql
此包为 https://github.com/childe/gohangout 项目的 用于mysql的outputs 插件。

# 特点
使用连接池，直接插入数据库

## TODO
批量插入，提升性能

# 使用方法

将 `mysql_output.go` 复制到 `gohangout` 主目录下面, 运行

```bash
go build -buildmode=plugin -o mysql_output.so mysql_output.go
```

将 `mysql_output.so` 路径作为 outputs

## gohangout 配置示例
所有参数字段名字都使用kafka-go原生的，所以和gohangout的kafka插件的配置名字有些不一样。主要是为了偷懒.

```yaml
inputs:
    - Stdin:
        codec: json

outputs:
    - Stdout:
        if:
            - '{{if .error}}y{{end}}'
    - '/Users/fiendhuang/program/my/gohangout/mysql_output.so':
        Host: '127.0.0.1'
        Port: 3306
        User: 'root'
        Password: '123'
        Database: 'test'
        TableKey: 'table'
        DefaultTable: 'test'
```
个别字段说明：   
- TableKey: 这个字段说明数据需要插入到哪个表里面
- DefaultTable：当数据里面没有TableKey字段的时候，数据插入到默认表名

# 数据说明
程序会遍历event，得到所有key-value的键值对，排除掉TableKey之后，所有key就是数据表的ColumnName。因此在数据插入之前需要保证表的数据结构的字段和json能够对得上。