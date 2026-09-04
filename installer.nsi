; Vizor Space — установщик NSIS, собирается прямо на Linux.
; Скрипт кладёт игру в Program Files и создаёт ярлык на Рабочем столе.

!define APP_NAME    "Vizor Space"
!define APP_VERSION  "1.0"
!define APP_EXE     "VizorSpace.exe"
!define APP_UNINST  "uninstall.exe"
!define MUI_ICON    "game.ico"
!define MUI_UNICON  "game.ico"
!include "MUI2.nsh"

Name "${APP_NAME}"
OutFile "VizorSpace-Setup.exe"
InstallDir "$PROGRAMFILES\VizorSpace"
InstallDirRegKey HKLM "Software\VizorSpace" "InstallDir"

RequestExecutionLevel admin
SetCompressor /SOLID lzma
Unicode true

XPStyle on

!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "Russian"

Section "-Install"
  SetOutPath "$INSTDIR"

  File "${APP_EXE}"

  WriteUninstaller "$INSTDIR\${APP_UNINST}"

  ; ярлык на Рабочем столе
  CreateShortCut "$DESKTOP\Vizor Space.lnk" "$INSTDIR\${APP_EXE}" "" "$INSTDIR\${APP_EXE}" 0

  ; ярлыки в меню «Пуск»
  CreateDirectory "$SMPROGRAMS\VizorSpace"
  CreateShortCut "$SMPROGRAMS\VizorSpace\Vizor Space.lnk" "$INSTDIR\${APP_EXE}" "" "$INSTDIR\${APP_EXE}" 0
  CreateShortCut "$SMPROGRAMS\VizorSpace\Uninstall.lnk" "$INSTDIR\${APP_UNINST}"

  ; записи для «Программы и компоненты»
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\VizorSpace" "DisplayName" "${APP_NAME}"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\VizorSpace" "DisplayVersion" "${APP_VERSION}"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\VizorSpace" "InstallLocation" "$INSTDIR"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\VizorSpace" "UninstallString" "$INSTDIR\${APP_UNINST}"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\VizorSpace" "Publisher" "Vizor Team"
  WriteRegStr HKLM "Software\VizorSpace" "InstallDir" "$INSTDIR"
SectionEnd

Section "Uninstall"
  ; аккуратно сметаем всё, что положили при установке
  Delete "$DESKTOP\Vizor Space.lnk"
  Delete "$SMPROGRAMS\VizorSpace\Vizor Space.lnk"
  Delete "$SMPROGRAMS\VizorSpace\Uninstall.lnk"
  RMDir "$SMPROGRAMS\VizorSpace"

  Delete "$INSTDIR\${APP_EXE}"
  Delete "$INSTDIR\${APP_UNINST}"
  RMDir "$INSTDIR"

  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\VizorSpace"
  DeleteRegKey HKLM "Software\VizorSpace"
SectionEnd