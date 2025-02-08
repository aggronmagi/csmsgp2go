MessagePack Code Generator
=======

fork from github.com/tinylib/msgp

compatible with https://github.com/MessagePack-CSharp/MessagePack-CSharp

兼容性修改:
1. 固定使用数组来序列化结构体. 索引从0开始. 索引号有间隔的,填充nil.
2. 不支持指针类型,否则会导致行为和csharp不一致
3. map的key类型支持数字和string
4. 非map的key之外的所有string, 允许为nil(go里面值为"")
5. map,slice,array 空值只设置数据头,不设置为nil
