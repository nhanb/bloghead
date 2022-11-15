package require Tk

wm title . "Create or Open?"

set OS [lindex $tcl_platform(os) 0]
if {$OS == "Windows"} {
    ttk::style theme use vista
} elseif {$OS == "Darwin"} {
    ttk::style theme use aqua
} else {
    ttk::style theme use clam
}

image create photo applicationIcon -data [
    binary decode base64 "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAQAAAAAYLlVAAAABGdBTUEAALGPC/xhBQAAACBjSFJNAAB6JgAAgIQAAPoAAACA6AAAdTAAAOpgAAA6mAAAF3CculE8AAAAAmJLR0QA/4ePzL8AAAAHdElNRQfmCwcRDCKC/5w4AAACo0lEQVRo3u2YzWsTQRiHn9nNNmms1BpUrDcj2tAWitSTCoKoN4vRg+DRQ/EmCAqCUP8BQS9ePFqkZ78QD0rx4yQ9tDSigkhKDxptlVqbZHfHg4na7FfS7I6K+5vbu++888zOvPMFsWL97xKtOBtatZ9eAGYpKuXUEwxxmWkqSCSS0bAiJ4IcNMMeIG8dow89iq75AGhJe4Dj9lFyaFE0HQhgj5NvbY6sRX59642+ef85YFL96RUZig+AyEsDgE4m2fYHAGSpBpKWVlTNE+X8jgFigBjgHwEI3I6D1CMW03KIfnpJsUyRGW3aXgmVUaR5VzuG/CqjAKKLE0xQwvrti8U819jVHd7u4Q4gMlxkzmGvl29cEd1RAtyn6Nl4vbwgGx1Ac+UN24OiR5sFWSbEen+XtWbBJPd4y0csNpLjAPtIufoNy0ucb7Mb3llQV0pjJ+OYrsPwlVzkAABJjVHKrgg3lAAAcG7VilAvC2z2jh7qJNSv8sjFvIGDigAskzHczo+HFAFA4jnTLuY9ygBMizsu5i1inSIA4BnSYeuUPeoACi4ABkllAKJExWn0vtqH/wcsllxsFS/3KDYjZyKWxZKXc/gAgg6H7XPikzqABM6Um63aygDkJpct/om3f/h/oM8R0+ShSgDnsvtKn1IGYGgccRivW6YygOogww2morjpVyNUgLTgQsMUlIzJxTbDtnAiGqHa4HdbN9ruV9MAeyk1eL2uPW37KKQh0JKc4S6ZVcY5RpgPqtn27VhPWDs4bJ9msOExs0Cel8H11wpwit0YdLHVypJxrP42t8TZ+jtj22rxbmgzw4gRZna1AFDhKSdFqpXozQ2B/2NDlTILTPGYB6nCipRNhWwu9A9G3dyPV68slvnCh473FTs4UqxYf6W+A90Ka5vyoUFeAAAAJXRFWHRkYXRlOmNyZWF0ZQAyMDIyLTExLTA3VDE2OjUyOjQ1KzAwOjAw84qPkAAAACV0RVh0ZGF0ZTptb2RpZnkAMjAyMi0xMS0wN1QxNjo1Mjo0NSswMDowMILXNywAAAAodEVYdGRhdGU6dGltZXN0YW1wADIwMjItMTEtMDdUMTc6MTI6MzQrMDA6MDAUP75ZAAAAAElFTkSuQmCC"
    # base64-encoded data generated with `base64 favicon.png`
]

wm iconphoto . -default applicationIcon

set types {
    {{Bloghead Files} {.bloghead}}
}

ttk::frame .c -padding "10"
ttk::label .c.label -text {Would you like to create a new blog, or open an existing one?}
ttk::button .c.createBtn -text "Create..." -padding 5 -command {
    set filename [tk_getSaveFile -title "Create" -filetypes $types]
    if {$filename != ""} {
        puts "create ${filename}"
        exit
    }
}
ttk::button .c.openBtn -text "Open..." -padding 5 -command {
    set filename [tk_getOpenFile -filetypes $types]
    if {$filename != ""} {
        puts "open ${filename}"
        exit
    }
}

grid .c -column 0 -row 0
grid .c.label -column 0 -row 0 -columnspan 2 -pady "0 10"
grid .c.createBtn -column 0 -row 1 -padx 10
grid .c.openBtn -column 1 -row 1 -padx 10

tk::PlaceWindow . center
