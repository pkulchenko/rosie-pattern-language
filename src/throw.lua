---- -*- Mode: Lua; -*-                                                                           
----
---- throw.lua
----
---- AUTHOR: Jamie A. Jennings

local co = require "coroutine"

function catch(f, ...)
   local c = co.create(f)
   local results = { co.resume(c, ...) }
   if results[1] then
      -- no errors
      return table.unpack(results, 2)
   end
   -- error occurred
   error(table.unpack(results, 2))
end

throw = co.yield

if false then
   tprint = table.print or function(t) for k,v in pairs(t) do print(k,v); end; end
      
   catch(tprint, {"a", "b"})
   --> prints the table, no errors
   print(pcall(catch, tprint, "asd", {"a", "b"}))
   --> error called, util.lua:106 bad argument #1 to 'pairs'
   print(catch(function() print("HELLO, WORLD!"); return 1,2,3; end))
   --> prints HELLO, WORLD!, and then 1  2  3
   print(catch(function() throw(5); print("HELLO, WORLD!"); return 1,2,3; end))
   --> prints 5
   print(nil==(catch(function() throw(); print("HELLO, WORLD!"); return 1,2,3; end)))
   --> prints true
   ok, msg = pcall(function() throw(4); end)
   assert(not ok)
   assert(msg=="attempt to yield from outside a coroutine")
   k = catch(pcall, function() throw(4); return 91921919; end)
   assert(k==4)
end