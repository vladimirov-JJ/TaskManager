package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Task struct {
	ID          uint
	Title       string
	Description string
}

func main() {
	a := app.New()
	a.Settings().SetTheme(theme.LightTheme())
	w := a.NewWindow("Task Manager")
	w.Resize(fyne.NewSize(500, 600))
	w.CenterOnScreen()

	var tasks []Task
	var createContent *fyne.Container
	var taskList *widget.List
	var taskContent *fyne.Container

	DB, _ := gorm.Open(sqlite.Open("todo.db"), &gorm.Config{})
	DB.AutoMigrate(&Task{})
	DB.Find(&tasks)

	noTaskLabel := canvas.NewText("No tasks", color.Black)

	if len(tasks) != 0 {
		noTaskLabel.Hide()
	}

	newTaskIcon, _ := fyne.LoadResourceFromPath("./icons/newTask.png")
	backIcon, _ := fyne.LoadResourceFromPath("./icons/back.png")
	saveIcon, _ := fyne.LoadResourceFromPath("./icons/save.png")
	editIcon, _ := fyne.LoadResourceFromPath("./icons/edit.png")
	deleteIcon, _ := fyne.LoadResourceFromPath("./icons/delete.png")

	taskBar := container.NewHBox(
		canvas.NewText("Your TASKS:", color.Black),
		layout.NewSpacer(),
		widget.NewButtonWithIcon("", newTaskIcon, func() {
			w.SetContent(createContent)
		}),
	)

	taskList = widget.NewList(
		func() int {
			return len(tasks)
		},

		func() fyne.CanvasObject {
			return widget.NewLabel("Default")
		},

		func(lii widget.ListItemID, co fyne.CanvasObject) {
			co.(*widget.Label).SetText(tasks[lii].Title)
		},
	)

	taskList.OnSelected = func(id widget.ListItemID) {
		detailsBar := container.NewHBox(
			canvas.NewText(
				fmt.Sprintf(
					"Details about %q", tasks[id].Title,
				),
				color.Black,
			),
			layout.NewSpacer(),
			widget.NewButtonWithIcon("", backIcon, func() {
				w.SetContent(taskContent)
				taskList.Unselect(id)
			}),
		)

		taskTitle := widget.NewLabel(tasks[id].Title)
		taskTitle.TextStyle = fyne.TextStyle{Bold: true}

		taskDescription := widget.NewLabel(tasks[id].Description)
		taskDescription.TextStyle = fyne.TextStyle{Italic: true}
		taskDescription.Wrapping = fyne.TextWrapBreak

		buttonsBox := container.NewHBox(

			//DELETE
			widget.NewButtonWithIcon(
				"",
				deleteIcon,

				func() {
					dialog.ShowConfirm("delete task", fmt.Sprintf("Уверен на счет удалени %s ?", tasks[id].Title), func(b bool) {
						if b {
							DB.Delete(&Task{}, "ID = ?", tasks[id].ID)
							DB.Find(&tasks)
							DB.Find(&tasks)

							if len(tasks) == 0 {
								noTaskLabel.Show()
							} else {
								noTaskLabel.Hide()
							}
						}

						w.SetContent(taskContent)
						taskList.Refresh()
						taskList.UnselectAll()
					}, w)
				},
			),

			widget.NewButtonWithIcon(
				"",
				editIcon,

				func() {
					editBar := container.NewHBox(
						canvas.NewText(
							fmt.Sprintf(
								"Editing %q", tasks[id].Title,
							),
							color.Black,
						),
						layout.NewSpacer(),
						widget.NewButtonWithIcon("", backIcon, func() {
							w.SetContent(taskContent)
							taskList.Unselect(id)
						}),
					)

					editTitle := widget.NewEntry()
					editTitle.SetText(tasks[id].Title)

					editDescription := widget.NewMultiLineEntry()
					editDescription.SetText(tasks[id].Description)

					editButton := widget.NewButtonWithIcon(
						"Save task",
						saveIcon,

						// EDIT FUNCTION
						func() {
							DB.Model(&Task{}).
								Where("ID = ?", tasks[id].ID).
								Updates(Task{
									Title:       editTitle.Text,
									Description: editDescription.Text,
								})

							DB.Find(&tasks)

							w.SetContent(taskContent)
							taskList.Refresh()
							taskList.UnselectAll()
						},
					)

					editContent := container.NewVBox(
						editBar,
						canvas.NewLine(color.Black),

						editTitle,
						editDescription,
						editButton,
					)

					w.SetContent(editContent)
				},
			),
		)

		detailsVBox := container.NewVBox(
			detailsBar,
			canvas.NewLine(color.Black),

			taskTitle,
			taskDescription,
			buttonsBox,
		)

		w.SetContent(detailsVBox)
		taskList.UnselectAll()

	}

	taskScroll := container.NewScroll(taskList)
	taskScroll.SetMinSize(fyne.NewSize(500, 500))

	taskContent = container.NewVBox(
		taskBar,
		canvas.NewLine(color.Black),
		noTaskLabel,
		taskScroll,
	)

	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Task title...")

	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetPlaceHolder("description text...")

	saveTaskButton := widget.NewButtonWithIcon("Save title", saveIcon, func() {
		task := Task{
			Title:       titleEntry.Text,
			Description: descriptionEntry.Text,
		}

		DB.Create(&task)
		DB.Find(&tasks)

		//вынести в отдельную функцию
		titleEntry.Text = ""
		titleEntry.Refresh()

		descriptionEntry.Text = ""
		descriptionEntry.Refresh()

		w.SetContent(taskContent)
		// taskList.Refresh()
		taskList.UnselectAll()

		if len(tasks) == 0 {
			noTaskLabel.Show()
		} else {
			noTaskLabel.Hide()
		}
	})

	createBar := container.NewHBox(
		canvas.NewText("Create new task", color.Black),
		layout.NewSpacer(),

		widget.NewButtonWithIcon("", backIcon, func() {
			titleEntry.Text = ""
			titleEntry.Refresh()
			descriptionEntry.Text = ""
			descriptionEntry.Refresh()

			w.SetContent(taskContent)
			taskList.UnselectAll()
		}),
	)

	createContent = container.NewVBox(
		createBar,
		canvas.NewLine(color.Black),

		container.NewVBox(
			titleEntry,
			descriptionEntry,
			saveTaskButton,
		),
	)

	w.SetContent(taskContent)
	w.Show()
	a.Run()
}
