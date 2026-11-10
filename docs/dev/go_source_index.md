# Go Source Index

Generated from current `.go` files using Go AST (`go/parser`, `go/ast`, `go/token`). Use this as a quick technical map for chat/session continuity.

## Repository Tree (Rooms summarized)

```text
WayFinder/
├── docs/
│   ├── autoMapper/
│   │   ├── DeveloperSpec.md
│   │   ├── RestartPrompt.md
│   │   └── Summary.md
│   ├── dev/
│   │   ├── scripts/
│   │   │   └── generate_go_source_index.go
│   │   ├── chat_context.txt
│   │   ├── go_source_index.md
│   │   ├── llm_context.txt
│   │   ├── MAINTENANCE.md
│   │   ├── MudOutput.txt
│   │   └── ToDo.md
│   ├── architecture.md
│   ├── discovery_model.md
│   └── mapper_rules.md
├── Rooms/
│   ├── AmongstTheRocks359.txt
│   ├── ...
│   └── WildernessTrailJunction385.txt
│   (465 files total)
├── wcs/
│   ├── telnet/
│   └── win32/
│       ├── live_windows.go
│       ├── proc_windows.go
│       └── shell_windows.go
├── weg/
│   ├── live_gateway.go
│   └── simulated_gateway.go
├── wmr/
│   ├── dev_view.go
│   ├── runtime.go
│   └── solver.go
├── wne/
│   ├── navigation_session.go
│   └── navigation_session_test.go
├── wtl/
│   ├── live_mode.go
│   └── simulated_mode.go
├── .gitattributes
├── .gitignore
├── AGENTS.md
├── go.mod
├── LICENSE
├── log.txt
├── main.go
├── README.md
├── WayFinder.code-workspace
└── WayFinder.exe
```

## Go File Tree

- `docs/dev/scripts/generate_go_source_index.go`
- `main.go`
- `wcs/win32/live_windows.go`
- `wcs/win32/proc_windows.go`
- `wcs/win32/shell_windows.go`
- `weg/live_gateway.go`
- `weg/simulated_gateway.go`
- `wmr/dev_view.go`
- `wmr/runtime.go`
- `wmr/solver.go`
- `wne/navigation_session.go`
- `wne/navigation_session_test.go`
- `wtl/live_mode.go`
- `wtl/simulated_mode.go`

## `docs/dev/scripts/generate_go_source_index.go`

Types:
- symbol (line 16)
- fileIndex (line 21)

Functions:
- main (line 28)
- collectGoFiles (line 58)
- indexFile (line 83)
- exprString (line 152)
- sortSymbols (line 160)
- buildRepositoryTree (line 169)
- buildTreeNode (line 179)
- summarizeRoomsTree (line 236)
- writeIndex (line 272)
- writeSymbols (line 301)
- fail (line 311)

Variables:
- (none)

## `main.go`

Types:
- (none)

Functions:
- main (line 15)

Variables:
- (none)

## `wcs/win32/live_windows.go`

Types:
- livePlayer (line 40)
- liveUpdate (line 42)
- liveShell (line 48)

Functions:
- RunLiveWCS (line 60)
- createLivePanels (line 74)
- layoutLivePanels (line 93)
- start (method on *liveShell) (line 109)
- postUpdate (method on *liveShell) (line 129)
- drainUpdates (method on *liveShell) (line 136)
- submitCommand (method on *liveShell) (line 163)
- appendTranscript (line 175)
- setWindowText (line 191)

Variables:
- liveWindow (line 57)

## `wcs/win32/proc_windows.go`

Types:
- (none)

Functions:
- (none)

Variables:
- user32 (line 6)
- kernel32 (line 7)
- gdi32 (line 8)
- procAdjustWindowRectEx (line 10)
- procCreateWindowExW (line 11)
- procDefWindowProcW (line 12)
- procDestroyWindow (line 13)
- procDispatchMessageW (line 14)
- procEnableWindow (line 15)
- procGetClientRect (line 16)
- procGetMessageW (line 17)
- procKillTimer (line 18)
- procLoadImageW (line 19)
- procMoveWindow (line 20)
- procPostQuitMessage (line 21)
- procRegisterClassExW (line 22)
- procSendMessageW (line 23)
- procSetFocus (line 24)
- procSetTimer (line 25)
- procShowWindow (line 26)
- procTranslateMessage (line 27)
- procUpdateWindow (line 28)
- procGetModuleHandleW (line 30)
- procGetStockObject (line 31)

## `wcs/win32/shell_windows.go`

Types:
- point (line 49)
- msg (line 54)
- rect (line 64)
- wndClassEx (line 71)

Functions:
- RunWCS (line 99)
- wndProc (line 179)
- createPanels (line 209)
- layoutPanels (line 227)
- moveWindow (line 273)
- createWindowEx (line 283)
- getModuleHandle (line 301)
- loadSystemResource (line 309)
- init (line 324)

Variables:
- hMainWnd (line 87)
- hWOV (line 88)
- hWIC (line 89)
- hWMP (line 90)
- hWOVLbl (line 91)
- hWICLbl (line 92)
- hWMPLbl (line 93)

## `weg/live_gateway.go`

Types:
- LiveText (line 9)
- LiveGateway (line 15)

Functions:
- NewLiveGateway (line 21)
- Write (method on *LiveGateway) (line 25)

Variables:
- (none)

## `weg/simulated_gateway.go`

Types:
- Result (line 21)
- SimulatedGateway (line 30)

Functions:
- NewSimulatedGateway (line 35)
- IngestRawText (method on *SimulatedGateway) (line 42)
- snapshot (method on *SimulatedGateway) (line 83)
- normalizeDirName (line 92)

Variables:
- (none)

## `wmr/dev_view.go`

Types:
- discoveredView (line 5)
- worldView (line 9)

Functions:
- PrintGrid10x10 (method on *Mapper) (line 13)
- PrintGrid10x10Discovered (method on *Mapper) (line 17)
- PrintRooms (method on *Mapper) (line 79)
- PrintRoomsDiscovered (method on *Mapper) (line 83)
- visibleExits (line 125)

Variables:
- (none)

## `wmr/runtime.go`

Types:
- RoomID (line 26)
- Room (line 28)
- Mapper (line 34)
- lockedAdjKey (line 44)
- ConstraintRelation (line 50)
- ConstraintSet (line 61)
- lockedAdjViolationError (line 143)
- collisionError (line 154)
- roomSnapshot (line 166)
- mapperSnapshot (line 172)
- plannedMove (line 1018)
- Topology (line 1655)

Functions:
- relationForKey (line 66)
- BuildConstraintSet (method on *Mapper) (line 84)
- Error (method on *lockedAdjViolationError) (line 149)
- Error (method on *collisionError) (line 161)
- SetUIOutput (line 182)
- uiPrint (line 190)
- uiPrintf (line 194)
- uiPrintln (line 198)
- setupLogging (line 202)
- NewMapper (line 217)
- BindTopology (method on *Mapper) (line 228)
- SetDebugWriter (method on *Mapper) (line 232)
- SetSolverProvider (method on *Mapper) (line 240)
- debugln (method on *Mapper) (line 248)
- debugf (method on *Mapper) (line 252)
- colName (line 256)
- cellLabel (line 258)
- normalizeDirName (line 271)
- dirDelta (line 297)
- getRoom (method on *Mapper) (line 322)
- clearOcc (method on *Mapper) (line 331)
- setOcc (method on *Mapper) (line 339)
- edgeAlignedAndOrdered (line 360)
- roomBetweenAxis (line 383)
- noRoomBetweenAxis (method on *Mapper) (line 401)
- refreshLockedAdjacencies (method on *Mapper) (line 421)
- validateLockedAdjacencies (method on *Mapper) (line 452)
- validateConstraintSet (method on *Mapper) (line 456)
- solverContext (method on *Mapper) (line 476)
- solver (method on *Mapper) (line 504)
- toSolverConstraintSet (method on *Mapper) (line 512)
- shiftWhere (method on *Mapper) (line 539)
- blockKey (line 616)
- formatBlockRooms (line 625)
- cloneBlock (line 634)
- printRejection (method on *Mapper) (line 642)
- destinationCandidateBlocks (method on *Mapper) (line 662)
- moveDestinationWithCandidates (method on *Mapper) (line 769)
- validateBlockMove (method on *Mapper) (line 862)
- moveBlock (method on *Mapper) (line 931)
- captureSnapshot (method on *Mapper) (line 948)
- restoreSnapshot (method on *Mapper) (line 973)
- stateSignature (method on *Mapper) (line 990)
- holeOpenNow (method on *Mapper) (line 1008)
- smallestBlocks (line 1025)
- planningBlocks (line 1041)
- blockHasID (line 1088)
- plannerDeltas (line 1093)
- planMakeRoomMultiStepDepth (method on *Mapper) (line 1120)
- planMakeRoomMultiStep (method on *Mapper) (line 1267)
- validateHoleOpens (method on *Mapper) (line 1316)
- makeRoom (method on *Mapper) (line 1355)
- rebuildDiscoveredLayout (method on *Mapper) (line 1491)
- Enter (method on *Mapper) (line 1524)
- enterIncremental (method on *Mapper) (line 1542)

Variables:
- uiOut (line 178)

## `wmr/solver.go`

Types:
- SolverRoomID (line 8)
- SolverLockedAdjKey (line 10)
- SolverConstraintRelation (line 16)
- SolverConstraintSet (line 27)
- SolverContext (line 32)
- SolverRoomState (line 42)
- RebuildRoomState (line 48)
- RebuildResult (line 54)
- SolverEngine (line 60)
- SolverProvider (line 65)
- LockedAdjViolationError (line 67)
- ConstraintSolver (line 78)

Functions:
- Error (method on *LockedAdjViolationError) (line 73)
- NewConstraintSolver (line 82)
- ValidateConstraintSet (method on *ConstraintSolver) (line 92)
- ComputeRebuildResult (method on *ConstraintSolver) (line 124)

Variables:
- DefaultSolverProvider (line 86)

## `wne/navigation_session.go`

Types:
- RoomID (line 5)
- Topology (line 7)
- World (line 13)
- Mapper (line 18)
- Discovery (line 23)
- Navigator (line 28)
- NavigationSession (line 38)

Functions:
- NewNavigationSession (line 45)
- CurrentRoom (method on *NavigationSession) (line 76)
- CurrentExits (method on *NavigationSession) (line 80)
- Move (method on *NavigationSession) (line 84)

Variables:
- (none)

## `wne/navigation_session_test.go`

Types:
- testWorld (line 5)
- testMapper (line 32)
- testDiscovery (line 46)

Functions:
- ExitsFrom (method on *testWorld) (line 9)
- Neighbors (method on *testWorld) (line 18)
- HasRoom (method on *testWorld) (line 27)
- BindTopology (method on *testMapper) (line 37)
- Enter (method on *testMapper) (line 41)
- Discover (method on *testDiscovery) (line 50)
- IsDiscovered (method on *testDiscovery) (line 54)
- buildWorld (line 59)
- TestNavigationSessionStartAndMove (line 69)
- TestNavigationSessionNoExit (line 96)

Variables:
- (none)

## `wtl/live_mode.go`

Types:
- (none)

Functions:
- StreamLiveOutput (line 16)
- PlayLive (line 30)

Variables:
- (none)

## `wtl/simulated_mode.go`

Types:
- World (line 58)
- DiscoveryState (line 113)
- wneTopologyAdapter (line 147)
- wneWorldAdapter (line 173)
- wneMapperAdapter (line 199)
- wneDiscoveryAdapter (line 211)
- localRoomPresentation (line 336)

Functions:
- uiPrint (line 28)
- uiPrintf (line 32)
- uiPrintln (line 36)
- setupLogging (line 40)
- ExitsFrom (method on *World) (line 63)
- Neighbors (method on *World) (line 75)
- HasRoom (method on *World) (line 88)
- ensureRoom (method on *World) (line 96)
- addExit (method on *World) (line 105)
- NewDiscoveryState (line 117)
- Discover (method on *DiscoveryState) (line 123)
- IsDiscovered (method on *DiscoveryState) (line 127)
- discoveredRoomIDs (line 135)
- toWNERoomID (line 144)
- fromWNERoomID (line 145)
- ExitsFrom (method on wneTopologyAdapter) (line 151)
- Neighbors (method on wneTopologyAdapter) (line 160)
- HasRoom (method on wneTopologyAdapter) (line 169)
- ExitsFrom (method on wneWorldAdapter) (line 177)
- Neighbors (method on wneWorldAdapter) (line 186)
- HasRoom (method on wneWorldAdapter) (line 195)
- BindTopology (method on wneMapperAdapter) (line 203)
- Enter (method on wneMapperAdapter) (line 207)
- Discover (method on wneDiscoveryAdapter) (line 215)
- LoadWorld (line 219)
- parseRoomFileIntoWorld (line 256)
- emitLocalPrompt (line 343)
- emitSimulatedRoomOutput (line 347)
- emitSimulatedMoveFailure (line 364)
- emitSimulatedSystemText (line 375)
- loadLocalRoomPresentation (line 381)
- normalizeDescription (line 432)
- wrapDescriptionLines (line 447)
- formatSimulatedExits (line 470)
- exitSortRank (line 495)
- formatExitDisplayName (line 518)
- toRoomIDExits (line 545)
- normalizeDirName (line 553)
- Run (line 578)

Variables:
- uiOut (line 26)
