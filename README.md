# TailorKV v1.0.0
A lightweight and customized KV cache.  
### How to use?
+ ##### Get server && client && config.xml of TailorKV.  
  + You can find ser & cli in /bin and config.xml is in /resource in this branch.
  + Or find them in the ```latest release```  .
  + Or you can clone this repo and use ```go build src/tailor_server/tailorServer.go``` and ```go build src/tailor_client/tailorCli.go```.  
+ ##### Create directory like this
  + ┏ /bin
  + ┃ &nbsp;&nbsp;&nbsp;┗ tailorServer.exe
  + ┗ /resource  
  + &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;┗ config.xml
+ ##### Start the server of TailorKV
  + ./tailorServer
+ ##### Use cli of TailorKV to connect TailorKV server
  + ./tailorCli -ip ``ip addr of server``
  + Such as ./tailorCli -ip 127.0.0.1
+ ##### Use instruction to control the TailorKV server 
  + ```set   [key] [val]```
  + ```setex [key] [val] [expiration]``` (expiration is millisecond)
  + ```setnx [key] [val]```
  + ```get   [key]```
  + ```del   [key]```
  + ```unlink [key]```
  + ```ttl   [key]```
  + ```incr  [key]```
  + ```incrby [key] [addition]``` (addition is integer)
  + ```cnt```
  + ```keys [regular expression]```
  + ```cls```
  + ```save```
  + ```save [filename]```
  + ```load```
  + ```load [filename]```
  + ```exit```
  + ```quit```