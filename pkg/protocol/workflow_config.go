package protocol

// Workflow也是一个CR资源，这里只需要定义WorkflowSpec即可
// 认为如果next为空串、或者是在NodesMap里都没有的名字，那么就是结束了！
// 如果什么都不做就返回，那么返回一个空串
type WorkflowSpec struct {
	EntryNode   string                  `yaml:"entryNode" json:"entryNode"`     // Workflow的入口节点，是一个WorkflowNode的name
	EntryParams map[string]interface{}  `yaml:"entryParams" json:"entryParams"` // Workflow的入口参数，请在yaml中按照json串的格式来描述它(也可以用yaml自己的键值对方式），例如"{"a": 1, "b": 2}"；这个参数只作为默认值，如果用户在触发时传入了参数，那么这里的参数会被覆盖
	Nodes       map[string]WorkflowNode `yaml:"nodes" json:"nodes"`             // Workflow的节点，key是节点的name，value是一个WorkflowNode
}

type WorkflowNode struct {
	Type          string            `yaml:"type" json:"type"` // 允许为func，choice，其中前两种需要有进一步的结构体解析
	FuncNodeRef   FuncNodeRefType   `yaml:"funcNodeRef" json:"funcNodeRef"`
	ChoiceNodeRef ChoiceNodeRefType `yaml:"choiceNodeRef" json:"choiceNodeRef"`
}

// 告知这些信息，才能调用一个函数
type FuncNodeRefType struct {
	Metadata MetadataType `yaml:"metadata" json:"metadata"`
	Next     string       `yaml:"next" json:"next"` // 标注执行完本函数后，跳转到哪一个Workflow节点，下一个节点允许为func, choice，如果保留空串则执行到这里就结束
}

// 接入编写go风格的逻辑表达式，在运行时会具体做eval操作
// 实现成从上到下的switch case逻辑
type ChoiceNodeRefType struct {
	Conditons []ChoiceConditionType `yaml:"conditions" json:"conditions"` // 只有使用切片才能保证从上到下遍历条件
}

type ChoiceConditionType struct {
	// 请用户自己保证每个condition会参与到的变量是OK的，防止发生错误，这里只做一个简单的描述
	Name       string   // 条件名称
	Variables  []string `yaml:"variables" json:"variables"`   // 参与到计算的变量列表，只做描述，不做检查，例如["a", "b"]；在运行时，仍然会根据上一个函数执行的结果来获取这些变量的值，并参与到以下Expression的计算中
	Expression string   `yaml:"expression" json:"expression"` // 可以被计算的逻辑表达式，例如"1 == 1"，如果其中有变量，那么在运行时会传递一个map键值对来进行eval代入计算
	Next       string   `yaml:"next" json:"next"`             // 如果这个条件为真，那么跳转到下一个Workflow节点，下一个节点允许为func, choice，如果为空，那么返回值是就是上一次执行FuncNode的计算结果；如果所有condition都不满足，那么也结束计算，返回上一次执行FuncNode的计算结果（自带的一个default case，当然也可以在最后写一个Expression恒为真的条件来手动设置default case的流向）
}
