## register
```
curl -X PUT --data @payload.json \
http://localhost:18550/v1/agent/service/register
```
```
curl -X PUT --data @payload2.json \
http://localhost:18550/v1/agent/service/register
```

## deregister
```
curl -X PUT \
    http://localhost:18550/v1/agent/service/deregister/web-01
```
```
curl -X PUT \
    http://localhost:18550/v1/agent/service/deregister/web-02
```

## discover
```
curl 'http://localhost:18500/v1/health/service/web?dc=dc1&passing=true'
```