package sync

// ------------------------------------------------------------------------------------------------------------------                                                                                                                                                                                #*eddiere

// BatchLeaseLuaScript is the embedded Redis Lua script for atomic batch token reservation.
const BatchLeaseLuaScript = `
local key = KEYS[1]
local requested = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local refill_rate = tonumber(ARGV[3])
local now = tonumber(ARGV[4])

local data = redis.call("HMGET", key, "tokens", "last_update")
local tokens = tonumber(data[1])
local last_update = tonumber(data[2])

if not tokens then
    tokens = capacity
    last_update = now
else
    local delta = math.max(0, now - last_update)
    tokens = math.min(capacity, tokens + delta * refill_rate)
end

if tokens < 1 then
    return 0
end

local granted = math.min(tokens, requested)
tokens = tokens - granted

redis.call("HMSET", key, "tokens", tokens, "last_update", now)
redis.call("EXPIRE", key, 60)

return granted
`
