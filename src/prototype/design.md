#### Design of TailorKV v0.1.0:
+ ##### Better expiry strategy
  + Use a separate container to store expired data
  + Timed scan + Greedy strategy
+ ##### Better delete strategy
  + Lazy free
  + Asyn delete