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
├── weg/
├── wmr/
│   ├── local_mode.go
│   ├── local_mud_output.go
│   ├── runtime.go
│   └── solver.go
├── wne/
│   ├── navigation_session.go
│   └── navigation_session_test.go
├── wtl/
├── .gitattributes
├── .gitignore
├── AGENTS.md
├── go.mod
├── LICENSE
├── log.txt
├── main.go
├── README.md
├── WayFinder.code-workspace
├── WayFinder.exe
├── WayFinder_refactor_step1_codex_instructions.md
└── WayFinder_refactor_step1_docs_task.md
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
- main (line 11)

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
- RunWCS (line 97)
- wndProc (line 166)
- createPanels (line 184)
- layoutPanels (line 202)
- moveWindow (line 244)
- createWindowEx (line 254)
- getModuleHandle (line 272)
- loadSystemResource (line 280)
- init (line 295)

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
- parseRoomFileIntoWorld (line 338)
- Run (line 422)

Variables:
- (none)

## `wmr/local_mud_output.go`

Types:
- localRoomPresentation (line 24)

Functions:
- emitLocalPrompt (line 31)
- emitSimulatedRoomOutput (line 35)
- emitSimulatedMoveFailure (line 52)
- emitSimulatedSystemText (line 63)
- loadLocalRoomPresentation (line 69)
- normalizeDescription (line 120)
- wrapDescriptionLines (line 135)
- formatSimulatedExits (line 158)
- exitSortRank (line 183)
- formatExitDisplayName (line 206)
- toRoomIDExits (line 233)

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
- plannedMove (line 1008)
- Topology (line 1645)

Functions:
- relationForKey (line 66)
- BuildConstraintSet (method on *Mapper) (line 84)
- Error (method on *lockedAdjViolationError) (line 149)
- Error (method on *collisionError) (line 161)
- uiPrint (line 180)
- uiPrintf (line 184)
- uiPrintln (line 188)
- setupLogging (line 192)
- NewMapper (line 207)
- BindTopology (method on *Mapper) (line 218)
- SetDebugWriter (method on *Mapper) (line 222)
- SetSolverProvider (method on *Mapper) (line 230)
- debugln (method on *Mapper) (line 238)
- debugf (method on *Mapper) (line 242)
- colName (line 246)
- cellLabel (line 248)
- normalizeDirName (line 261)
- dirDelta (line 287)
- getRoom (method on *Mapper) (line 312)
- clearOcc (method on *Mapper) (line 321)
- setOcc (method on *Mapper) (line 329)
- edgeAlignedAndOrdered (line 350)
- roomBetweenAxis (line 373)
- noRoomBetweenAxis (method on *Mapper) (line 391)
- refreshLockedAdjacencies (method on *Mapper) (line 411)
- validateLockedAdjacencies (method on *Mapper) (line 442)
- validateConstraintSet (method on *Mapper) (line 446)
- solverContext (method on *Mapper) (line 466)
- solver (method on *Mapper) (line 494)
- toSolverConstraintSet (method on *Mapper) (line 502)
- shiftWhere (method on *Mapper) (line 529)
- blockKey (line 606)
- formatBlockRooms (line 615)
- cloneBlock (line 624)
- printRejection (method on *Mapper) (line 632)
- destinationCandidateBlocks (method on *Mapper) (line 652)
- moveDestinationWithCandidates (method on *Mapper) (line 759)
- validateBlockMove (method on *Mapper) (line 852)
- moveBlock (method on *Mapper) (line 921)
- captureSnapshot (method on *Mapper) (line 938)
- restoreSnapshot (method on *Mapper) (line 963)
- stateSignature (method on *Mapper) (line 980)
- holeOpenNow (method on *Mapper) (line 998)
- smallestBlocks (line 1015)
- planningBlocks (line 1031)
- blockHasID (line 1078)
- plannerDeltas (line 1083)
- planMakeRoomMultiStepDepth (method on *Mapper) (line 1110)
- planMakeRoomMultiStep (method on *Mapper) (line 1257)
- validateHoleOpens (method on *Mapper) (line 1306)
- makeRoom (method on *Mapper) (line 1345)
- rebuildDiscoveredLayout (method on *Mapper) (line 1481)
- Enter (method on *Mapper) (line 1514)
- enterIncremental (method on *Mapper) (line 1532)

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
