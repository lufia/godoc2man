// graphite-metric-test is a verifier for Graphite Plaintext Protocol format.
//
// Usage
//
//	graphite-metric-test [-f rule] [file ...]
//
// graphite-metric-test reads the rules, reads metrics from stdin by default, and verify them.
// Then it ends up reports missing metrics, unexpected metrics or out of range metrics.
//
// Options
//
// The -f option is a file contains rules with metric path patterns and metric value ranges.
//
// The Rules
//
// The rule described in the rule file each lines is a pair of metric path pattern and value range.
//
//	// comment
//	local.random.diceroll	>0, <=6	 // v > 0 && v <= 6
//	local.thermal.*.temp	<=100000 // wildcard (* or #) matches any stem in the path
//	~local.network.tx.bytes	>0 // path starting with ~ is optional
//	local.uptime // no range; it checks path existence but the value is not checked
//
// If you want to check metrics with OR condition, you can put multiple lines with same path pattern.
//
//	local.signal.level		>=0, <2
//	local.signal.level		>=3, <5
//
// The Operators
//
// The operators are '<=', '<', '>=' and '>'.
package main
