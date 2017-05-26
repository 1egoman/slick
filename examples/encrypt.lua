local http = require("http")

--[[
A command to send an encrypted username to a user via keybase.

What this command does:
- Fetches a public key of a user from keybase.
- Imports the key into gpg
- Encrypts and armors the message with gpg.
- Sends a message to the current connection and channel with the encrypted message.

Try out this script by running `/require path/to/examples/encrypt.lua`,  and running `/encrypt`!

--]]
command("encrypt",
	"Send an encrypted message to another user via keybase",
	"<keybase username> <message>",
	function(args)
		username = args[2]
		message = args[3]
		if not username or not message then
			error("ERROR: please specify a username and message! /encrypt <username> <message>")
		else
			print("Fetching key for "..username.."...")

			response, err = http.request("GET", "https://keybase.io/"..username.."/key.asc")
			if err then
				error(err)
			else
				print("Got key for "..username..". Importing into gpg...")
				cert = response.body
				message = "foo"

				_, err = shell("which", "gpg")
				if err then
					error("Please install gpg and make sure it's in your path!")
					return
				end

				-- First, import the certificate
				_, err = shell("sh", "-c", "echo '"..cert.."' | gpg --import")
				if err then
					error(err)
					return
				end

				-- Then, encrypt the thing
				print("Encrypting...")
				output, err = shell(
					"sh", "-c",
					"echo '"..message.."' | gpg --encrypt --armor --recipient keybase.io/"..username.." --trust-model always"
				)

				if err then
					error(err)
					return
				end

        clear()
				sendmessage("```\n"..output.."\n```")
			end
		end
	end
)
