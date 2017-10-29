```
title: The big door
required: [{name: key, value: silver}]

```

{{ if .Locked }}

# You shall not pass...

Find the key to advance further and beyond.

{{ else }}

# The door is now open

Nothing can't stop you. See you in the /lake.

{{ end }}