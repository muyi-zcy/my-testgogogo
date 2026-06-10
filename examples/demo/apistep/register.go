// Package apistep 封装 demo 示例的 API 调用步骤，供单接口测试和 Flow 流程复用。
//
// 通过 blank import 触发 demoauth Provider 的 init() 注册。
package apistep

import _ "github.com/muyi-zcy/my-testgogogo/examples/demo/internal/demoauth"
