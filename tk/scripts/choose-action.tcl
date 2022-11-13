package require Tk
wm title . "Create or Open?"
tk::PlaceWindow . center
ttk::style theme use clam

set types {
    {{Bloghead Files} {.bloghead}}
}

ttk::frame .c -padding "10"
ttk::label .c.label -text {Would you like to create a new blog, or open an existing one?}
ttk::button .c.createBtn -text "Create..." -padding 5 -command {
    set filename [tk_getSaveFile -title "Create" -filetypes $types]
    if {$filename != ""} {
        puts "create"
        puts $filename
        exit
    }
}
ttk::button .c.openBtn -text "Open..." -padding 5 -command {
    set filename [tk_getOpenFile -filetypes $types]
    if {$filename != ""} {
        puts "open"
        puts $filename
        exit
    }
}

grid .c -column 0 -row 0
grid .c.label -column 0 -row 0 -columnspan 2 -pady "0 10"
grid .c.createBtn -column 0 -row 1 -padx 10
grid .c.openBtn -column 1 -row 1 -padx 10
