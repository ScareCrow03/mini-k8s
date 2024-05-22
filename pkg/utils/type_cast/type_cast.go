package type_cast

import (
	"encoding/json"
	"fmt"
)

// 传递源接口src（泛型），转换到目标类型的对象（传递一个指针进来）dst
func GetObjectFromInterface(src interface{}, dst interface{}) error {
	bytes, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("failed to marshal src: %w", err)
	}
	err = json.Unmarshal(bytes, dst)
	if err != nil {
		return fmt.Errorf("failed to unmarshal to dst: %w", err)
	}
	return nil
}
