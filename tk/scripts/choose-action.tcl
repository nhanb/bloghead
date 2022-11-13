package require Tk
wm title . "Create or Open?"
tk::PlaceWindow . center
ttk::style theme use clam

ttk::frame .c -padding "10"

ttk::label .c.label -text {Would you like to create a new blog, or open an existing one?}
ttk::button .c.createBtn -text "Create..." -command {puts "create"; exit} -padding 5
ttk::button .c.openBtn -text "Open..." -command {puts "open"; exit} -padding 5

grid .c -column 0 -row 0
grid .c.label -column 0 -row 0 -columnspan 2 -pady "0 10"
grid .c.createBtn -column 0 -row 1 -padx 10
grid .c.openBtn -column 1 -row 1 -padx 10

wm protocol . WM_DELETE_WINDOW {
    puts "cancel"
    exit
}
