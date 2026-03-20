package functions

// RegisterBuiltins registers all built-in DSL functions into the evaluator.
func RegisterBuiltins(r Registrar) {
	r.RegisterFunction(&NoopFunction{})
	r.RegisterFunction(&AddFunction{})
	r.RegisterFunction(&SubFunction{})
	r.RegisterFunction(&MulFunction{})
	r.RegisterFunction(&DivFunction{})

	r.RegisterFunction(&EqFunction{})
	r.RegisterFunction(&NeqFunction{})
	r.RegisterFunction(&GtFunction{})
	r.RegisterFunction(&GteFunction{})
	r.RegisterFunction(&LtFunction{})
	r.RegisterFunction(&LteFunction{})

	r.RegisterFunction(&AndFunction{})
	r.RegisterFunction(&OrFunction{})
	r.RegisterFunction(&NotFunction{})

	r.RegisterFunction(&ListFunction{})
	r.RegisterFunction(&LengthFunction{})
	r.RegisterFunction(&MapFunction{})
	r.RegisterFunction(&FilterFunction{})
	r.RegisterFunction(&ConcatFunction{})
	r.RegisterFunction(&JoinStringFunction{})
	r.RegisterFunction(&InStringFunction{})
	r.RegisterFunction(&ObjGetByPathFunction{})
	r.RegisterFunction(&ObjGetValFunction{})
	r.RegisterFunction(&ObjSetByPathFunction{})
	r.RegisterFunction(&SetFunction{})
	r.RegisterFunction(&ReduceFunction{})
	r.RegisterFunction(&RandFunction{})
	r.RegisterFunction(&PluginCardFunction{})
	r.RegisterFunction(&OneCardFunction{})
	r.RegisterFunction(&ForEachKeyValFunction{})
	r.RegisterFunction(&VarFunction{})
	r.RegisterFunction(&PrevFunction{})
	r.RegisterFunction(&CondFunction{})
	r.RegisterFunction(&WhileFunction{})
	r.RegisterFunction(&MoveCardFunction{})
}
