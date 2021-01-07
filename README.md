# secretary

## secretary-lab

### Getting Started

1. prepare `config.yaml`

```bash
$ cp sample-config.yaml config.yaml
```

```yaml
# config.yaml
users:
  shanpu: # username
    slack_id: "Uxxxxx"   # user's slack id
    slack_channel: "nil"    # slack channel
    secretary_name: "sample"   # bot name
    secretary_icon: ":slack:"  # bot icon(your slack-workspace's custom icon name)
    secretary_coming_msg: ""   # bot message when you come
    secretary_goodbye_msg: ""  # bot message when you leave
  rinse:
    slack_id: ...
    ...

```

2. set environment vars in `docker-compose.yml`
   - `LAB_SLACK_TOKEN`: slack bot token 
   - `LAB_SLACK_COMING_CHANNEL`: slack channel ID to report entry

3. run
```bash
$ docker-compose up -d
```
