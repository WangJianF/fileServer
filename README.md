# fileServer
游戏热更新，日志服务器

在客户端中上传日志:
```javascript
let xhr = new XMLHttpRequest()
xhr.onreadystatechange = () => {
    if (xhr.readyState === 4) {
        if (xhr.status >= 200 && xhr.status < 400) {
            console.log('upload success');
        } else {
            console.log('upload failure');
        }
    }
};
xhr.open('POST', 'http://localhost:8888/uploadLog/', true)
xhr.setRequestHeader('Content-type', 'application/json');
xhr.send(JSON.stringify({
    RecordTime: Math.Floor(Date.now()/1000),
    UserID: 123456,
    UserName: 'jugg',
    Desc: 'upload log example',
    Content: '...' //log content
}));

```

上传热更新包:

要求为zip格式，且命名为xxx时间戳.zip, 热更新url：http://localhost:8888/hotupdate/target


