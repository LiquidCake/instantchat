## profile backend:
1. add import

```_ "net/http/pprof"```

2. add separate http server that will catch up import automatically

```go http.ListenAndServe("0.0.0.0:6060", nil)```

3. map/open port 6060

4. open page e.g.

http://127.0.0.1:6060/debug/pprof/
   
5. download 'profile', open it with 

```go tool pprof .\profile```
   
6. analyze profile via pprof. E.g. make sure https://graphviz.org/ is installed and export with pprof 'svg' command 
