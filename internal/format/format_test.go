package format

import "testing"

func TestSourceFormatsConsistently(t *testing.T) {
	input := "fn main(){let value=1+2*3;if value>0 {return value!!;} else {return 0;}}"
	formatted, err := Source(input)
	if err != nil {
		t.Fatalf("Source returned error: %v", err)
	}
	const expected = `fn main() {
    let value = 1 + 2 * 3;
    if value > 0 {
        return value!!;
    } else {
        return 0;
    }
}
`
	if formatted != expected {
		t.Fatalf("unexpected formatted output:\n--- got ---\n%q\n--- want ---\n%q", formatted, expected)
	}
}
