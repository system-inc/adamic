exec(open('/tmp/defend-interface-cast/run.py').read().split('plans=[')[0])
plans=[('D05','internal/lower/class_inheritance.go','if !viewSite(node) || node.Kind == ast.KindParenthesizedExpression {','if !viewSite(node) || node.Kind == ast.KindParenthesizedExpression || (node.Kind == ast.KindNewExpression && l.iteratorMember(l.checker.GetTypeAtLocation(node)) != nil) {','Return early for iterable constructor assignments before nominal-view verification')]
s=open('/tmp/defend-interface-cast/run.py').read();exec(s[s.index('for ident,file,old,new,why in plans:'):])
