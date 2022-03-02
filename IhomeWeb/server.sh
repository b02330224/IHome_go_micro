sudo redis-server ./conf/redis.conf

fdfs_trackerd  /home/gopath/src/sss/Ihomeweb/conf/tracker.conf restart
#启动storaged

fdfs_storaged  /home/gopath/src/sss/Ihomeweb/conf/storage.conf restart

nginx