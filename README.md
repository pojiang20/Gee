# Gee
简单学习gin框架

#### 上下文设计理念
针对使用场景，封装`*http.Request`和`http.ResponseWriter`的方法，简化相关接口的调用，只是设计 `Context` 的原因之一。对于框架来说，还需要支撑额外的功能。例如，将来解析动态路由`/hello/:name`，参数`:name`的值放在哪呢？再比如，框架需要支持中间件，那中间件产生的信息放在哪呢？`Context` 随着每一个请求的出现而产生，请求的结束而销毁，和当前请求强相关的信息都应由 `Context` 承载。因此，设计 `Context` 结构，扩展性和复杂性留在了内部，而对外简化了接口。路由的处理函数，以及将要实现的中间件，参数都统一使用 `Context` 实例， `Context` 就像一次会话的百宝箱，可以找到任何东西。
#### 接口型函数
定义函数类型`HandlerFunc`
```go
type HandlerFunc func(*Context)
```
作为传入接口之后
```go
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	...
}
```
可以以匿名函数的形式传入，非常好用。
```go
r.GET("/", func(c *gee.Context) {
    c.HTML(http.StatusOK, "<h1>Hello Gee</ht>")
})
```
#### 前缀树
路由的路径之间存在前缀关系，可以使用前缀树来实现。查询（`O(n*m)->O(m)`）和存储都更优。把动态区域`:xx`作为特殊节点`:`存储即可。
```go

type node struct {
	value string
	next  map[string]*node
}

var rootNode = &node{
	value: "/",
}

func insert(path string) {
	routeBlock := strings.Split(path, "/")

	p := rootNode
	for _, v := range routeBlock {
		//未找到
		if _, ok := p.next[v]; !ok {
			//所有:lang\:day等形式都视为:，作为动态区域的标记
			if strings.HasPrefix(v, ":") {
				v = ":"
			}
			//判断是否为动态区域
			p.next[v] = &node{
				value: v,
			}
		}
		p = p.next[v]
	}
}

func match(path string) *node {
	routeBlock := strings.Split(path, "/")

	p := rootNode
	for _, part := range routeBlock {
		//未找到
		if nextNode, ok := p.next[part]; !ok {
			if nextNode.value == ":" {
				part = ":"
			}
			return nil
		}
		p = p.next[part]
	}
	return p
}
```
#### 分组路由
在添加路由过程中，增加公共前缀。
```go
func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}
```
### 中间件
有时候需要对执行函数进行前置后置处理，这样就形成了类似`a1(b1(myFunc())b2)a2`的嵌套。
#### 注册与实现
使用`r.Use(gee.Logger())`来注册中间件，具体则是向`middlewares`数组中追加内容。
```go
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}
```
先追加中间件，最后再追加执行的操作。`c.Next()`从头开始遍历，先执行中间件最后执行目的操作。
```go
func (r *router) handle(c *Context) {
    ...
	c.handlers = append(c.handlers, r.handlers[key])
	...
    c.Next()
}
```
这里维护了`index`来记录调用的层数，以此达到嵌套调用的效果。
```go
func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}
```
这样只要在`next()`的前后调用函数，就可以起到执行前置后置函数的效果。
```go
// 前置函数
t := time.Now()
//内嵌函数
c.Next()
//后置函数
log.Printf("[%d] %s in %v", c.StatusCode, c.Req.RequestURI, time.Since(t))
```

[参考](https://geektutu.com/post/gee.html)