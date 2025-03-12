package main

import (
	"fmt"
	"os"

	"github.com/Monsler/gandon/gandonc"
	Qt "github.com/mappu/miqt/qt6"
)

func main() {
	Qt.NewQApplication(os.Args)
	//var gandon gandonc.GanDecryptor
	Handle := Qt.NewQWidget(nil)
	Handle.SetWindowTitle("Gandon")
	Handle.SetMinimumSize2(400, 230)

	EditText1 := Qt.NewQLineEdit(Handle)
	EditText1.SetPlaceholderText("Путь к гандону")
	ButtonSelectGandonPath := Qt.NewQPushButton(Handle)
	ButtonSelectGandonPath.SetText("Выбрать путь к гандону")
	EditText2 := Qt.NewQLineEdit(Handle)
	EditText2.SetPlaceholderText("Папка вывода")
	ButtonSelectExportPath := Qt.NewQPushButton(Handle)
	ButtonSelectExportPath.SetText("Выбрать путь для вывода")
	ButtonProcess := Qt.NewQPushButton(Handle)
	ButtonProcess.SetText("Дешифровать")

	Handle.OnResizeEvent(func(super func(event *Qt.QResizeEvent), event *Qt.QResizeEvent) {
		EditText1.SetGeometry(10, 10, Handle.Width()-20, 30)
		ButtonSelectGandonPath.SetGeometry(10, 50, Handle.Width()-20, 30)
		EditText2.SetGeometry(10, 90, Handle.Width()-20, 30)
		ButtonSelectExportPath.SetGeometry(10, 130, Handle.Width()-20, 30)
		ButtonProcess.SetGeometry(10, Handle.Height()-40, Handle.Width()-20, 30)
	})

	ButtonSelectGandonPath.OnClicked(func() {
		evt := Qt.NewQFileDialog(nil)
		evt.Show()
		evt.OnFileSelected(func(file string) {
			EditText1.SetText(file)
		})
	})

	ButtonSelectExportPath.OnClicked(func() {
		evt := Qt.NewQFileDialog(nil)
		evt.SetFileMode(Qt.QFileDialog__Directory)
		evt.Show()
		evt.OnFileSelected(func(file string) {
			EditText2.SetText(file)
		})

	})

	ButtonProcess.OnClicked(func() {
		if EditText1.Text() != "" && EditText2.Text() != "" {
			gandon, err := gandonc.NewGanDecryptor(EditText1.Text(), EditText2.Text())
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			gandon.Process()
			dlg := Qt.NewQMessageBox(nil)
			dlg.SetText(fmt.Sprintf("Done! Saved to %s", EditText2.Text()))
			dlg.Show()
		}
	})
	Handle.Show()
	Qt.QApplication_Exec()
}
