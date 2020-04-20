#### Design of TailorKV prototype:
+ ##### operation of cache
  + ```set``` &nbsp;&nbsp;&nbsp; &lt;key&gt; &lt;val&gt;
  + ```setex``` &lt;ex&gt; &lt;key&gt; &lt;val&gt;
  + ```setnx``` &lt;key&gt; &lt;val&gt;
  + ```get``` &nbsp;&nbsp;&nbsp; &lt;key&gt;
  + ```del``` &nbsp;&nbsp;&nbsp; &lt;key&gt;
  + ```incr``` &nbsp;&nbsp; &lt;key&gt;
  + ```incrby``` &lt;key&gt; &lt;num&gt;
+ ##### Special features
  + custom daemon task
  + concurrent
  + custom daemon cleaner thread
  + ...
+ ##### Enhancement in the future
  + Only one mutex is used to synchronizing the ```cache```, which is not efficient enough.
  + All the ```data structures``` are from ```Go SDK```, I need to do some improvements according to the actual situation.
  + ```Persistence``` needs to be improved.
  + build a terminal for TailorKV
  + ...
  