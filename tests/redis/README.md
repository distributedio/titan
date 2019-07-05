Description:
* You need to install tcl 8.5 or higher(same as redis 5.0.5)
* After testing, you need to cleanup the test data by sending flushdb commands or manually deleting it before you can test other types.

command:
```
tclsh test_helper.tcl --single <test file path> --host <ip> --port <port> --auth <token>
```
* single # Specify a test file in ./unit or ./unit/type directory, eg : --single unit/type/string
* host   # the server's ip which we need to test 
* port   # the server's port which we need to test 
* auth   # the token


