# Shell Agent
Shell Agent, is a program installed on remote host, help  you to execute shell command on the remote host. 

This agent can also help to transport file to/from remote host.

# Install 
```
go build
./shell_agent 
```
Now you can visit the agent via 8080 port.

The usage is simple:
```
Shell Agent.

        Usage:
        shell_agent [--cnf=<path>] [--addr=<addr>]
        shell_agent -h | --help
        shell_agent --version

        Options:
        --cnf=<path>  config file path [default: ].
        --addr=<addr>  config file path [default: :8080].
```



# Execute command
There are two ways to execute command on remote host: sync(is default) and async.
## sync
Type the following curl request will hang and wait for `sleep 5` to finish.
```
curl -d '{"cmd":"sleep 5 && echo test sleep"}' http://127.0.0.1:8080/api/v1/cmd/run
{"errno":0,"error":"succeed","data":{"id":"bda8616a-0179-4c54-468b-918e15112006","status":"finished","error":"","cmd":"sleep 600 \u0026\u0026 echo test sleep","env":null,"stdout":"test sleep\n","stderr":"","exit_code":0,"pid":10266,"create_time":"2018-02-25T19:01:32.870915989+08:00","finish_time":"2018-02-25T19:11:32.875958476+08:00"}}
```

Take a careful look at the response:
```
{
    "errno": 0, 
    "error": "succeed", 
    "data": {
        "id": "bda8616a-0179-4c54-468b-918e15112006", 
        "status": "finished", 
        "error": "", 
        "cmd": "sleep 600 && echo test sleep", 
        "env": null, 
        "stdout": "test sleep\n", 
        "stderr": "", 
        "exit_code": 0, 
        "pid": 10266, 
        "create_time": "2018-02-25T19:01:32.870915989+08:00", 
        "finish_time": "2018-02-25T19:11:32.875958476+08:00"
    }
}
```
The returned http response contain: 
* id: UUID of the job .
* status: Status of the job, maybe running, finished(the command exited with zero exit code), failed(the command failed to start, or be killed, or exited with non-zero exit code)
* error: The reason why the job failed.
* stdout: Stdout of the command.
* stderr: Stderr of the command.
* exit_code: Exit code of the command.


## async
Type the following curl request will not hang and return immediately.
The returned http response only contain the job id, which can be use to query job progress.
```
curl -d '{"cmd":"sleep 600 && echo test sleep", "async":true}' http://127.0.0.1:8080/api/v1/cmd/run  
{"errno":0,"error":"succeed","data":{"id":"3dcb8bb9-5aab-4a5c-7575-fa11294d2dff","create_time":"2018-02-25T19:38:38.539287299+08:00"}}
```
