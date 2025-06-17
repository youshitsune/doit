# DoIt

## Description

Server for managing tasks

## Getting Started 

### Binary
You can just install binary for your system from releases.

Write a config on in the same directory as executable.
```yaml
port: "3333"
username: "default"
password: "default"
```

Change default credentials, don't be stupid!

Just start it.
```
./doit
```

### Docker
You need to build docker image (I'll maybe setup GitHub Action).

```
git clone https://github.com/youshitsune/doit
cd doit/
earthly +docker
```

Write a config.yaml
```yaml
port: "3333"
username: "default"
password: "default"
```

Docker container run command:
```
docker run -idt -p <port_on_your_system>:<port_set_in_config> -v <path_to_config>:/etc/doit/config.yaml --name doit doit
```

## API Routes
- /new    (POST) Make new task, arguments: task name and tag name
- /list   (GET)  Return a list of all task
- /delete (POST) Delete a task, arguments: id of task
- /change (POST) Change a task state, arguments: id of task
- /rename (POST) Rename a task, arguments: id of task and new task name
