dir := ""

@default:
    just --list

bar foo:
    echo {{ if foo == "bar" { "hello" } else { "goodbye" } }}

# analyze TYPE:
#     {{
#         dir = if TYPE == "oreo"{
#     "pkg"
#     }else if TYPE == "benchmark"{
#     "benchmarks"
#     }else{
#     ""
#     }
#     }}
#     fd -e go -E '*_test.go' . {{dir}} | xargs tokei