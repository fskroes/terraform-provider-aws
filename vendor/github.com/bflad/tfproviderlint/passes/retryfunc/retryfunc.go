package retryfunc

import (
	"go/ast"
	"reflect"

	"github.com/bflad/tfproviderlint/helper/terraformtype"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name: "retryfunc",
	Doc:  "find github.com/hashicorp/terraform-plugin-sdk/helper/resource RetryFunc declarations for later passes",
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	Run:        run,
	ResultType: reflect.TypeOf([]*terraformtype.HelperResourceRetryFuncInfo{}),
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.FuncLit)(nil),
	}
	var result []*terraformtype.HelperResourceRetryFuncInfo

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		funcDecl, funcDeclOk := n.(*ast.FuncDecl)
		funcLit, funcLitOk := n.(*ast.FuncLit)

		var funcType *ast.FuncType

		if funcDeclOk && funcDecl != nil {
			funcType = funcDecl.Type
		} else if funcLitOk && funcLit != nil {
			funcType = funcLit.Type
		} else {
			return
		}

		params := funcType.Params

		if params != nil && len(params.List) != 0 {
			return
		}

		results := funcType.Results

		if results == nil || len(results.List) != 1 {
			return
		}

		if !terraformtype.IsHelperResourceTypeRetryError(pass.TypesInfo.TypeOf(results.List[0].Type)) {
			return
		}

		result = append(result, terraformtype.NewHelperResourceRetryFuncInfo(funcDecl, funcLit, pass.TypesInfo))
	})

	return result, nil
}
