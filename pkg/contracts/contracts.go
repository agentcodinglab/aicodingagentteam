// Package contracts provides OpenAPI contract parsing and cross-validation
// between frontend API calls and backend definitions.
package contracts

// Contract is a parsed OpenAPI contract.
type Contract struct {
	OpenAPI string
	Paths   map[string]PathItem // path pattern -> methods
}

// PathItem holds the operations for a single path.
type PathItem struct {
	Get    *Operation
	Post   *Operation
	Put    *Operation
	Delete *Operation
	Patch  *Operation
}

// Operation describes a single API operation.
type Operation struct {
	OperationID string
	Method      string
	Path        string
	Summary     string
}

// CallSite is a detected frontend API call (fetch/axios) to be cross-checked.
type CallSite struct {
	File   string
	Method string
	Path   string
	Line   int
}

// Mismatch is a contract violation found during cross-validation.
type Mismatch struct {
	Call     CallSite
	Reason   string
	Severity string // blocking / advisory
}

// CrossCheck compares frontend call sites against the OpenAPI contract.
// Returns blocking mismatches that should halt delivery.
func CrossCheck(contract *Contract, calls []CallSite) []Mismatch {
	var out []Mismatch
	for _, c := range calls {
		item, ok := contract.Paths[matchPath(contract, c.Path)]
		if !ok {
			out = append(out, Mismatch{Call: c, Reason: "path not in contract: " + c.Path, Severity: "blocking"})
			continue
		}
		if !methodExists(item, c.Method) {
			out = append(out, Mismatch{Call: c, Reason: "method not allowed: " + c.Method, Severity: "blocking"})
		}
	}
	return out
}

func matchPath(c *Contract, p string) string { return p }

func methodExists(item PathItem, method string) bool {
	switch method {
	case "GET":
		return item.Get != nil
	case "POST":
		return item.Post != nil
	case "PUT":
		return item.Put != nil
	case "DELETE":
		return item.Delete != nil
	case "PATCH":
		return item.Patch != nil
	}
	return false
}
