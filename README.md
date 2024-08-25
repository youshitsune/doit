# DoIt

## Description

Server for managing tasks

## Getting Started 

### Binary
You can just install binary for your system from releases.

Write a config on /etc/doit/config.yaml
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
- /new Make new task, arguments: task name and tag name
- /list Return a list of all task
- /done Mark a task as done, arguments: id of task
- /delete Delete a task, arguments: id of task
- /reset Reset a task state, arguments: id of task
- /rename Rename a task, arguments: id of task and new task name
- /getnote Get a note of task, arguments: id of task
- /newnote Make new note or update existing, arguments: id of task and note
- /deletenote Delete a note, arguments: id of task
- /edittag Edit a tag, arguments: id of task and new tag name
