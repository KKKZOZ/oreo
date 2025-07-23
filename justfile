@default:
    just --list

@push:
    jj pre-commit
    jj bookmark set main -r @
    jj git push

