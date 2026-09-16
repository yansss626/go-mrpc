## 请求/响应

统一定义为结构体：

```go
type <method>Request struct { // 请求
    <param_name> <type_name>
    ...
}

type <method>Response struct { // 响应
     <param_name> <type_name>
    ...  
}
```

## 头部

用于解决 TCP 的分包和问题，以及长度校验。

```
crc32(uni32) -- requestID(uint64) -- length(uint32)
```

## 请求格式

```json
{
  "service": "HelloService",
  "method": "SayHello",
  "request": {
    "msg": "hello",
    "length": 5
  }
}
```

## 响应格式

```
{
  "code": 0,
  "message": "",
  "response": {
    "message": "olleh"
  }
} // success

{
  "code": 1001,
  "message": "method not found",
  "response": null
}
```

