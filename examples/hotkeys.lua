-- "c"hannel "g"eneral
keymap("cg", function()
  err = Pick("Test Team", "general")
  if err then
    error(err)
  end
end)

-- "c"hannel "r"eneral
keymap("cr", function()
  err = Pick("Test Team", "random")
  if err then
    error(err)
  end
end)
