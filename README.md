# goskills

## Подключение к проекту

### git clone
- клонируйте этот репозиторий в любой каталог на вашем ПК
- подключите к вашей среде разработки на примере OpenCode и macOS. 
  Отредактируйте файл ~/.config/opencode/opencode.jsonc следующим образом:
```json
{
  "$schema": "https://opencode.ai/config.json",
  "skills": [
    "~/.config/opencode/skills",
    "~/Documents/GitHub/goskills/skills/"
  ]
}
```
- скиллы, подключенные через opencode.json, новый проект увидит автоматически. 
  OpenCode каждый раз сам обходит каталоги и показывает агенту список доступных навыков, 
  а агент уже решает, какой из них загрузить.
- обновление версии через `git pull`