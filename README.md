# testtool-golang-ginkgo

## 用例属性

### 用例标签

参考[ginkgo文档](https://onsi.github.io/ginkgo/#spec-labels)，插件支持加载用例中声明的标签`Label`并以键值对的格式设置在用例属性`Attributes`内，键值对格式为`label: "[label01, label02]"`，其中键为固定的`label`，值为一个json序列化的字符串列表，列表中每个元素表示用例的一个标签。

例如，存在用例声明如下:

```golang
It("is labelled", Label("first label"), Label("second label"), func() { ... })
```

则用例对应的用例属性为:

`label: "[first label, second label]"`

加载或执行时可以在`TestSelector`中声明对应的标签以实现过滤，例如:

```shell
# 加载当前用例库中标签中包含label01的用例
solarctl load -t ".?label=label01"

# 执行当前用例库中标签中包含label01的用例
solarctl run -t ".?label=label01"
```