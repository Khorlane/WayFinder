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
├── solver/
│   └── solver.go
├── wcs/
│   ├── telnet/
│   └── win32/
│       ├── proc_windows.go
│       └── shell_windows.go
├── wmr/
│   ├── local_mode.go
│   ├── local_mud_output.go
│   └── runtime.go
├── wne/
│   ├── navigation_session.go
│   └── navigation_session_test.go
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
- `solver/solver.go`
- `wcs/win32/proc_windows.go`
- `wcs/win32/shell_windows.go`
- `wmr/local_mode.go`
- `wmr/local_mud_output.go`
- `wmr/runtime.go`
- `wne/navigation_session.go`
- `wne/navigation_session_test.go`

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
- main (line 9)

Variables:
- (none)

## `solver/solver.go`

Types:
- RoomID (line 8)
- LockedAdjKey (line 10)
- ConstraintRelation (line 16)
- ConstraintSet (line 27)
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

## `wcs/win32/proc_windows.go`

Types:
- (none)

Functions:
- (none)

Variables:
- user32 (line 6)
- kernel32 (line 7)
- procAdjustWindowRectEx (line 9)
- procCreateWindowExW (line 10)
- procDefWindowProcW (line 11)
- procDestroyWindow (line 12)
- procDispatchMessageW (line 13)
- procGetClientRect (line 14)
- procGetMessageW (line 15)
- procLoadImageW (line 16)
- procMoveWindow (line 17)
- procPostQuitMessage (line 18)
- procRegisterClassExW (line 19)
- procShowWindow (line 20)
- procTranslateMessage (line 21)
- procUpdateWindow (line 22)
- procGetModuleHandleW (line 24)

## `wcs/win32/shell_windows.go`

Types:
- point (line 48)
- msg (line 53)
- rect (line 62)
- wndClassEx (line 69)

Functions:
- RunWCS (line 94)
- wndProc (line 163)
- createPanels (line 181)
- layoutPanels (line 199)
- moveWindow (line 241)
- createWindowEx (line 251)
- getModuleHandle (line 269)
- loadSystemResource (line 277)
- init (line 292)

Variables:
- hMainWnd (line 85)
- hWOV (line 86)
- hWIC (line 87)
- hWMP (line 88)
- hWOVLbl (line 89)
- hWICLbl (line 90)
- hWMPLbl (line 91)

## `wmr/local_mode.go`

Types:
- World (line 127)
- DiscoveryState (line 165)
- wneTopologyAdapter (line 209)
- wneWorldAdapter (line 235)
- wneMapperAdapter (line 261)
- wneDiscoveryAdapter (line 273)

Functions:
- PrintGrid10x10 (method on *Mapper) (line 15)
- PrintGrid10x10Discovered (method on *Mapper) (line 19)
- PrintRooms (method on *Mapper) (line 81)
- PrintRoomsDiscovered (method on *Mapper) (line 85)
- ExitsFrom (method on *World) (line 132)
- Neighbors (method on *World) (line 144)
- HasRoom (method on *World) (line 157)
- NewDiscoveryState (line 169)
- Discover (method on *DiscoveryState) (line 175)
- IsDiscovered (method on *DiscoveryState) (line 179)
- visibleExits (line 187)
- discoveredRoomIDs (line 197)
- toWNERoomID (line 206)
- fromWNERoomID (line 207)
- ExitsFrom (method on wneTopologyAdapter) (line 213)
- Neighbors (method on wneTopologyAdapter) (line 222)
- HasRoom (method on wneTopologyAdapter) (line 231)
- ExitsFrom (method on wneWorldAdapter) (line 239)
- Neighbors (method on wneWorldAdapter) (line 248)
- HasRoom (method on wneWorldAdapter) (line 257)
- BindTopology (method on wneMapperAdapter) (line 265)
- Enter (method on wneMapperAdapter) (line 269)
- Discover (method on wneDiscoveryAdapter) (line 277)
- LoadWorld (line 281)
- ensureRoom (method on *World) (line 318)
- addExit (method on *World) (line 327)
- parseRoomFileIntoWorld (line 335)
- Run (line 416)

Variables:
- (none)

## `wmr/local_mud_output.go`

Types:
- localRoomPresentation (line 21)

Functions:
- emitLocalPrompt (line 28)
- emitSimulatedRoomOutput (line 32)
- emitSimulatedMoveFailure (line 49)
- emitSimulatedSystemText (line 60)
- loadLocalRoomPresentation (line 66)
- normalizeDescription (line 117)
- wrapDescriptionLines (line 132)
- formatSimulatedExits (line 155)
- exitSortRank (line 180)
- formatExitDisplayName (line 203)
- toRoomIDExits (line 230)

Variables:
- (none)

## `wmr/runtime.go`

Types:
- RoomID (line 25)
- Room (line 27)
- Mapper (line 33)
- lockedAdjKey (line 43)
- ConstraintRelation (line 49)
- ConstraintSet (line 60)
- lockedAdjViolationError (line 142)
- collisionError (line 153)
- roomSnapshot (line 165)
- mapperSnapshot (line 171)
- plannedMove (line 1007)
- Topology (line 1644)

Functions:
- relationForKey (line 65)
- BuildConstraintSet (method on *Mapper) (line 83)
- Error (method on *lockedAdjViolationError) (line 148)
- Error (method on *collisionError) (line 160)
- uiPrint (line 179)
- uiPrintf (line 183)
- uiPrintln (line 187)
- setupLogging (line 191)
- NewMapper (line 206)
- BindTopology (method on *Mapper) (line 217)
- SetDebugWriter (method on *Mapper) (line 221)
- SetSolverProvider (method on *Mapper) (line 229)
- debugln (method on *Mapper) (line 237)
- debugf (method on *Mapper) (line 241)
- colName (line 245)
- cellLabel (line 247)
- normalizeDirName (line 260)
- dirDelta (line 286)
- getRoom (method on *Mapper) (line 311)
- clearOcc (method on *Mapper) (line 320)
- setOcc (method on *Mapper) (line 328)
- edgeAlignedAndOrdered (line 349)
- roomBetweenAxis (line 372)
- noRoomBetweenAxis (method on *Mapper) (line 390)
- refreshLockedAdjacencies (method on *Mapper) (line 410)
- validateLockedAdjacencies (method on *Mapper) (line 441)
- validateConstraintSet (method on *Mapper) (line 445)
- solverContext (method on *Mapper) (line 465)
- solver (method on *Mapper) (line 493)
- toSolverConstraintSet (method on *Mapper) (line 501)
- shiftWhere (method on *Mapper) (line 528)
- blockKey (line 605)
- formatBlockRooms (line 614)
- cloneBlock (line 623)
- printRejection (method on *Mapper) (line 631)
- destinationCandidateBlocks (method on *Mapper) (line 651)
- moveDestinationWithCandidates (method on *Mapper) (line 758)
- validateBlockMove (method on *Mapper) (line 851)
- moveBlock (method on *Mapper) (line 920)
- captureSnapshot (method on *Mapper) (line 937)
- restoreSnapshot (method on *Mapper) (line 962)
- stateSignature (method on *Mapper) (line 979)
- holeOpenNow (method on *Mapper) (line 997)
- smallestBlocks (line 1014)
- planningBlocks (line 1030)
- blockHasID (line 1077)
- plannerDeltas (line 1082)
- planMakeRoomMultiStepDepth (method on *Mapper) (line 1109)
- planMakeRoomMultiStep (method on *Mapper) (line 1256)
- validateHoleOpens (method on *Mapper) (line 1305)
- makeRoom (method on *Mapper) (line 1344)
- rebuildDiscoveredLayout (method on *Mapper) (line 1480)
- Enter (method on *Mapper) (line 1513)
- enterIncremental (method on *Mapper) (line 1531)

Variables:
- uiOut (line 177)

## `wne/navigation_session.go`

Types:
- RoomID (line 5)
- Topology (line 7)
- World (line 13)
- Mapper (line 18)
- Discovery (line 23)
- Navigator (line 28)
- NavigationSession (line 36)

Functions:
- NewNavigationSession (line 43)
- CurrentRoom (method on *NavigationSession) (line 74)
- CurrentExits (method on *NavigationSession) (line 78)
- Move (method on *NavigationSession) (line 82)

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
