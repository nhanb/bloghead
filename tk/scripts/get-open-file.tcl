package require Tk
wm withdraw .
ttk::style theme use clam
set types {
    {{Bloghead Files} {.bloghead}}
}
set filename [tk_getOpenFile -filetypes $types]
puts $filename
exit
