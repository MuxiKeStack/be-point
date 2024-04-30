---
--- Generated by EmmyLua(https://github.com/EmmyLua)
--- Created by xishan.
--- DateTime: 2024/4/30 19:20
---

local key = KEYS[1]
local exists = redis.call("EXISTS", key)
if exists == 1 then
    redis.call("INCR", key)
    return 1
else
    return 0
end