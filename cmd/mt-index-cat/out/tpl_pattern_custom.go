package out

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
)

type bucket struct {
	chance int
	fn     func(in string) string
}

func PatternCustomUsage(indent string) {
	fmt.Printf("%spatternCustom: transforms a graphite.style.metric.name into a pattern with wildcards inserted according to rules provided:\n", indent)
	fmt.Printf("%s               patternCustom <chance> <operation>[ <chance> <operation>...]\n", indent)
	fmt.Printf("%s               the chances need to add up to 100\n", indent)
	fmt.Printf("%s               operation is one of:\n", indent)
	fmt.Printf("%s                 * pass        (passthrough)\n", indent)
	fmt.Printf("%s                 * <digit>rcnw (replace a randomly chosen sequence of <digit (0-9)> consecutive nodes with wildcards\n", indent)
	fmt.Printf("%s                 * <digit>rccw (replace a randomly chosen sequence of <digit (0-9)> consecutive characters with wildcards\n", indent)
	fmt.Printf("%s               example: {{.Name | patternCustom 15 \"pass\" 40 \"1rcnw\" 15 \"2rcnw\" 10 \"3rcnw\" 10 \"3rccw\" 10 \"2rccw\"}}\\n\n", indent)
}

// percentage chance, and function
func patternCustom(in ...interface{}) string {
	usage := func() {
		PatternCustomUsage("")
		os.Exit(-1)
	}

	// one or more of "<chance> <operation>" followed by an input string at the end.
	if len(in) < 3 || len(in)%2 != 1 {
		usage()
	}
	input, ok := in[len(in)-1].(string)
	if !ok {
		usage()
	}
	var buckets []bucket
	var sum int
	for i := 0; i < len(in)-2; i += 2 {
		chance, ok := in[i].(int)
		if !ok {
			usage()
		}
		patt, ok := in[i+1].(string)
		if !ok {
			usage()
		}
		if patt == "pass" {
			sum += chance
			buckets = append(buckets, bucket{
				chance: chance,
				fn:     Passthrough,
			})
			continue
		}
		if patt[0] < '0' || patt[0] > '9' {
			usage()
		}
		num := int(patt[0] - '0') // parse ascii number to int
		if patt[1:] != "rcnw" && patt[1:] != "rccw" {
			usage()
		}
		var fn func(in string) string
		if patt[1:] == "rcnw" {
			fn = ReplaceRandomConsecutiveNodesWildcard(num)
		} else {
			fn = ReplaceRandomConsecutiveCharsWildcard(num)
		}

		sum += chance
		buckets = append(buckets, bucket{
			chance: chance,
			fn:     fn,
		})
	}
	if sum != 100 {
		usage()
	}
	pos := rand.Intn(100)
	sum = 0
	for _, b := range buckets {
		if pos < sum+b.chance {
			return b.fn(input)
		}
		sum += b.chance

	}
	panic("should never happen")
	return "foo"
}

func Passthrough(in string) string {
	return in
}

// ReplaceRandomConsecutiveNodesWildcard returns a function that will replace num consecutive random nodes with wildcards
// the implementation is rather naive and can be optimized
func ReplaceRandomConsecutiveNodesWildcard(num int) func(in string) string {
	return func(in string) string {
		parts := strings.Split(in, ".")
		if len(parts) < num {
			log.Fatalf("metric %q has not enough nodes to replace %d nodes", in, num)
		}
		pos := rand.Intn(len(parts) - num + 1)
		for i := pos; i < pos+num; i++ {
			parts[pos] = "*"
		}
		return strings.Join(parts, ".")
	}
}

// ReplaceRandomConsecutiveCharWildcard returns a function that will replace num consecutive random characters with wildcards
// note: it's also possible to produce patterns that won't match anything (if '.' was taken out)
// note: assumes string is long enough
func ReplaceRandomConsecutiveCharsWildcard(num int) func(in string) string {
	return func(in string) string {
		// visualizing the logic here... imagine num=2
		// abcd.fghi.klmn.pqrs.uvwx.z  (len=26)
		// ^ possible vals for pos ^
		// 0 .................... 24
		// let's say it's 12, so:
		//             ^
		// abcd.fghi.kl + ** + .qrs.uvwx.z
		if len(in) < num {
			log.Fatalf("metric %q not long enough to replace %d characters", in, num)
		}
		pos := rand.Intn(len(in) - num + 1)
		return in[0:pos] + strings.Repeat("*", num) + in[pos+num:]
	}
}
