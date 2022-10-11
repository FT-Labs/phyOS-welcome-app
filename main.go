package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/uitools"
	"github.com/therecipe/qt/widgets"
)

func NewInstallerApp() *widgets.QWidget {
    file := core.NewQFile2(":/qml/main.ui")
    file.Open(core.QIODevice__ReadOnly)
    installerWidget := uitools.NewQUiLoader(nil).Load(file, nil)
    file.Close()

    var (
        l_checkpass = widgets.NewQLabelFromPointer(installerWidget.FindChild("l_checkpass", core.Qt__FindChildrenRecursively).Pointer())
        pb_next = widgets.NewQPushButtonFromPointer(installerWidget.FindChild("pb_next", core.Qt__FindChildrenRecursively).Pointer())
        c_isencrypt = widgets.NewQCheckBoxFromPointer(installerWidget.FindChild("c_isencrypt", core.Qt__FindChildrenRecursively).Pointer())
        le_pass1 = widgets.NewQLineEditFromPointer(installerWidget.FindChild("le_pass1", core.Qt__FindChildrenRecursively).Pointer())
        le_pass2 = widgets.NewQLineEditFromPointer(installerWidget.FindChild("le_pass2", core.Qt__FindChildrenRecursively).Pointer())
        cb_swap = widgets.NewQComboBoxFromPointer(installerWidget.FindChild("cb_swap", core.Qt__FindChildrenRecursively).Pointer())
    )

    pb_next.ConnectClicked(func(check bool){
        f, err := os.Create("/tmp/phyinstall.conf")

        if err != nil {
            panic("Can't create settings file")
        }
        defer f.Close()

        cmd := "chmod +x /tmp/phyinstall.conf"
        exec.Command("/bin/bash", "-c", cmd).Run()

        f.WriteString("#!/bin/bash\n")

        if c_isencrypt.IsChecked() {
            f.WriteString("lvm_device=phyoscrypt\n")
            f.WriteString("is_encrypt=true\n")
            passw := fmt.Sprintf("pass=%s\n", le_pass1.Text())
            f.WriteString(passw)
        } else {
            cmd = "cp -f /etc/calamares/modules/partition-noluks2.conf /etc/calamares/modules/partition.conf"
            exec.Command("/bin/bash", "-c", cmd).Run()
            f.WriteString("lvm_device=phyos\n")
            f.WriteString("is_encrypt=false\n")
            f.WriteString("pass=\n")
        }
        swp := fmt.Sprintf("swap=%s\n", cb_swap.CurrentText())
        f.WriteString(swp)
        NewWelcomeApp().Show()
        installerWidget.Close()
    })

    c_isencrypt.ConnectClicked(func(checked bool){
        if checked {
            pb_next.SetEnabled(false)
            le_pass1.SetReadOnly(false)
            le_pass2.SetReadOnly(false)
        } else {
            le_pass1.Clear()
            le_pass2.Clear()
            l_checkpass.Clear()
            le_pass1.SetReadOnly(true)
            le_pass2.SetReadOnly(true)
            pb_next.SetEnabled(true)
        }
    })

    le_pass1.ConnectTextChanged(func(text string) {
        if text == "" {
            l_checkpass.SetText("Password can't be empty!")
        } else if strings.Contains(text, " ") {
            l_checkpass.SetText("Password contains space!")
        } else if text != le_pass2.Text() {
            l_checkpass.SetText("Passwords don't match")
        } else {
            l_checkpass.SetText("Password confirmed")
        }

        if l_checkpass.Text() == "Password confirmed" {
            pb_next.SetEnabled(true)
        } else {
            pb_next.SetEnabled(false)
        }
    })

    le_pass2.ConnectTextChanged(func(text string) {
        if text == "" {
            l_checkpass.SetText("Password can't be empty!")
        } else if strings.Contains(text, " ")  {
            l_checkpass.SetText("Password contains space!")
        } else if text != le_pass1.Text() {
            l_checkpass.SetText("Passwords don't match")
        } else {
            l_checkpass.SetText("Password confirmed")
        }

        if l_checkpass.Text() == "Password confirmed" {
            pb_next.SetEnabled(true)
        } else {
            pb_next.SetEnabled(false)
        }
    })

    return installerWidget
}

func NewWelcomeApp() *widgets.QWidget {
    file := core.NewQFile2(":/qml/install.ui")
    file.Open(core.QIODevice__ReadOnly)
    mainWidget := uitools.NewQUiLoader(nil).Load(file, nil)
    file.Close()

    var (
        pb_internet = widgets.NewQPushButtonFromPointer(mainWidget.FindChild("pb_internet", core.Qt__FindChildrenRecursively).Pointer())
        pb_pacman_keyring = widgets.NewQPushButtonFromPointer(mainWidget.FindChild("pb_pacman_keyring", core.Qt__FindChildrenRecursively).Pointer())
        pb_installer = widgets.NewQPushButtonFromPointer(mainWidget.FindChild("pb_installer", core.Qt__FindChildrenRecursively).Pointer())
        pb_installer_offline = widgets.NewQPushButtonFromPointer(mainWidget.FindChild("pb_installer_offline", core.Qt__FindChildrenRecursively).Pointer())
        pb_quit = widgets.NewQPushButtonFromPointer(mainWidget.FindChild("pb_quit", core.Qt__FindChildrenRecursively).Pointer())
    )

    pb_pacman_keyring.ConnectClicked(OnMirrorClicked)

    pb_quit.ConnectClicked(func(check bool) {
        os.Exit(0)
    })

    pb_installer.ConnectClicked(func(check bool) {
        cala := exec.Command("/usr/bin/calamares_polkit", "-d")
        cala.Start()
        cmd := "xdotool key Super_L+f"
        exec.Command("/bin/bash", "-c", cmd).Run()

    })

    pb_installer_offline.ConnectClicked(func(check bool){
        moveConf := exec.Command("sudo", "mv", "-f", "/etc/calamares/settings-default.conf",
                                 "/etc/calamares/settings.conf")
        moveConf.Start()
        cala := exec.Command("/usr/bin/calamares_polkit", "-d")
        cala.Start()
        cmd := "xdotool key Super_L+f"
        exec.Command("/bin/bash", "-c", cmd).Run()
    })

    pb_internet.ConnectClicked(func(check bool) {
        nmtuiCmd := exec.Command("st", "-e", "nmtui")
        nmtuiCmd.Start()
    })


    return mainWidget
}

func OnMirrorClicked(check bool) {
    go RefreshMirrors()
}

func RefreshMirrors() {
    exec.Command("dunstify", "-a", "center", "Refreshing Mirrors, Please Wait!", "&").Start()
    mirrorCmd := exec.Command("pkexec", "/usr/bin/reflector", "--age", "6", "--latest",
    "21", "--fastest", "21", "--threads", "21", "--sort", "rate", "--protocol", "https", "--save", "/etc/pacman.d/mirrorlist")

    err := mirrorCmd.Run()

    if err != nil {
        _, err := exec.Command("dunstify", "-u", "critical", "-a", "top-center", "Error occurred renewing keyring!").Output()
        fmt.Fprintf(os.Stderr, fmt.Sprintf("Error occurred: %s", err))
        return
    }

    keysCmd := exec.Command("dunstify", "-a", "center", "Mirrors refreshed succesfully!")
    _, err = keysCmd.Output()

    if err != nil {
        fmt.Fprintf(os.Stderr, fmt.Sprintf("Error occurred: %s", err))
    }
}

func main() {
    app := widgets.NewQApplication(len(os.Args), os.Args)
    app.SetDesktopSettingsAware(true)
    NewInstallerApp().Show()
    widgets.QApplication_Exec()
}
