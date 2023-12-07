## 1.数据的版本信息
### Type
string
### key -> value
index:{dc}:{service} -> uint64

## 2. 实际的数据
### Type
string

### key -> value
data:{dc}:{service}:{index}

"github.com/hashicorp/consul/api"
value 是 []*api.ServiceEntry的json字符串