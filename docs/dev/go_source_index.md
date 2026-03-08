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
│       ├── proc_windows.go
│       └── shell_windows.go
├── wmr/
│   ├── local_mode.go
│   ├── local_mud_output.go
│   ├── runtime.go
│   └── solver.go
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
- `wcs/win32/proc_windows.go`
- `wcs/win32/shell_windows.go`
- `wmr/local_mode.go`
- `wmr/local_mud_output.go`
- `wmr/runtime.go`
- `wmr/solver.go`
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
- RoomID (line 23)
- Room (line 25)
- Mapper (line 31)
- lockedAdjKey (line 41)
- ConstraintRelation (line 47)
- ConstraintSet (line 58)
- lockedAdjViolationError (line 140)
- collisionError (line 151)
- roomSnapshot (line 163)
- mapperSnapshot (line 169)
- plannedMove (line 1005)
- Topology (line 1642)

Functions:
- relationForKey (line 63)
- BuildConstraintSet (method on *Mapper) (line 81)
- Error (method on *lockedAdjViolationError) (line 146)
- Error (method on *collisionError) (line 158)
- uiPrint (line 177)
- uiPrintf (line 181)
- uiPrintln (line 185)
- setupLogging (line 189)
- NewMapper (line 204)
- BindTopology (method on *Mapper) (line 215)
- SetDebugWriter (method on *Mapper) (line 219)
- SetSolverProvider (method on *Mapper) (line 227)
- debugln (method on *Mapper) (line 235)
- debugf (method on *Mapper) (line 239)
- colName (line 243)
- cellLabel (line 245)
- normalizeDirName (line 258)
- dirDelta (line 284)
- getRoom (method on *Mapper) (line 309)
- clearOcc (method on *Mapper) (line 318)
- setOcc (method on *Mapper) (line 326)
- edgeAlignedAndOrdered (line 347)
- roomBetweenAxis (line 370)
- noRoomBetweenAxis (method on *Mapper) (line 388)
- refreshLockedAdjacencies (method on *Mapper) (line 408)
- validateLockedAdjacencies (method on *Mapper) (line 439)
- validateConstraintSet (method on *Mapper) (line 443)
- solverContext (method on *Mapper) (line 463)
- solver (method on *Mapper) (line 491)
- toSolverConstraintSet (method on *Mapper) (line 499)
- shiftWhere (method on *Mapper) (line 526)
- blockKey (line 603)
- formatBlockRooms (line 612)
- cloneBlock (line 621)
- printRejection (method on *Mapper) (line 629)
- destinationCandidateBlocks (method on *Mapper) (line 649)
- moveDestinationWithCandidates (method on *Mapper) (line 756)
- validateBlockMove (method on *Mapper) (line 849)
- moveBlock (method on *Mapper) (line 918)
- captureSnapshot (method on *Mapper) (line 935)
- restoreSnapshot (method on *Mapper) (line 960)
- stateSignature (method on *Mapper) (line 977)
- holeOpenNow (method on *Mapper) (line 995)
- smallestBlocks (line 1012)
- planningBlocks (line 1028)
- blockHasID (line 1075)
- plannerDeltas (line 1080)
- planMakeRoomMultiStepDepth (method on *Mapper) (line 1107)
- planMakeRoomMultiStep (method on *Mapper) (line 1254)
- validateHoleOpens (method on *Mapper) (line 1303)
- makeRoom (method on *Mapper) (line 1342)
- rebuildDiscoveredLayout (method on *Mapper) (line 1478)
- Enter (method on *Mapper) (line 1511)
- enterIncremental (method on *Mapper) (line 1529)

Variables:
- uiOut (line 175)

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
