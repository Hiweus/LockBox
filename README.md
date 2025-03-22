# LockBox

CLI tool to manage `totp` credentials :lock:

## TODO :bulb:
- [X] Refactor encryption module to Vault module
  - [X] Keep the password in memory
  - [X] Accept only input and give output as byte array
- [X] Make login prompt (generate .sync-lb)
  - [X] Break the method when has to receive user input (token/username) from when just access the filesystem
- [X] Perform sync operation on background
  - [X] Add metadata on .credentials-lb.json to represent sync status
    - [X] created_at
    - [X] updated_at
    - [X] synced_at or just synced
- [ ] Add export feature
  - [ ] JSON
  - [ ] Generate QR code to import into another tools

