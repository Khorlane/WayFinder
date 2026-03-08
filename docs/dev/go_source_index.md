# Go Source Index

Generated from current `.go` files using Go AST (`go/parser`, `go/ast`, `go/token`). Use this as a quick technical map for chat/session continuity.

## Go File Tree

- `docs/dev/scripts/generate_go_source_index.go`
- `internal/harness/navigation_session.go`
- `internal/harness/navigation_session_test.go`
- `internal/harness/runtime.go`
- `internal/wcs/win32/proc_windows.go`
- `internal/wcs/win32/shell_windows.go`
- `main.go`
- `solver/solver.go`

## `docs/dev/scripts/generate_go_source_index.go`

Types:
- symbol (line 16)
- fileIndex (line 21)

Functions:
- main (line 28)
- collectGoFiles (line 53)
- indexFile (line 78)
- exprString (line 147)
- sortSymbols (line 155)
- writeIndex (line 164)
- writeSymbols (line 187)
- fail (line 197)

Variables:
- (none)

## `internal/harness/navigation_session.go`

Types:
- Navigator (line 6)
- NavigationSession (line 16)

Functions:
- NewNavigationSession (line 23)
- CurrentRoom (method on *NavigationSession) (line 49)
- CurrentExits (method on *NavigationSession) (line 53)
- Move (method on *NavigationSession) (line 57)
- Mapper (method on *NavigationSession) (line 70)
- Discovery (method on *NavigationSession) (line 74)

Variables:
- (none)

## `internal/harness/navigation_session_test.go`

Types:
- (none)

Functions:
- testWorld (line 5)
- TestNavigationSessionStartAndMove (line 15)
- TestNavigationSessionNoExit (line 42)

Variables:
- (none)

## `internal/harness/runtime.go`

Types:
- RoomID (line 28)
- Room (line 30)
- Mapper (line 36)
- lockedAdjKey (line 46)
- ConstraintRelation (line 52)
- ConstraintSet (line 63)
- lockedAdjViolationError (line 145)
- collisionError (line 156)
- roomSnapshot (line 168)
- mapperSnapshot (line 174)
- plannedMove (line 1010)
- Topology (line 1761)
- World (line 1767)
- DiscoveryState (line 1805)

Functions:
- relationForKey (line 68)
- BuildConstraintSet (method on *Mapper) (line 86)
- Error (method on *lockedAdjViolationError) (line 151)
- Error (method on *collisionError) (line 163)
- uiPrint (line 182)
- uiPrintf (line 186)
- uiPrintln (line 190)
- setupLogging (line 194)
- NewMapper (line 209)
- BindTopology (method on *Mapper) (line 220)
- SetDebugWriter (method on *Mapper) (line 224)
- SetSolverProvider (method on *Mapper) (line 232)
- debugln (method on *Mapper) (line 240)
- debugf (method on *Mapper) (line 244)
- colName (line 248)
- cellLabel (line 250)
- normalizeDirName (line 263)
- dirDelta (line 289)
- getRoom (method on *Mapper) (line 314)
- clearOcc (method on *Mapper) (line 323)
- setOcc (method on *Mapper) (line 331)
- edgeAlignedAndOrdered (line 352)
- roomBetweenAxis (line 375)
- noRoomBetweenAxis (method on *Mapper) (line 393)
- refreshLockedAdjacencies (method on *Mapper) (line 413)
- validateLockedAdjacencies (method on *Mapper) (line 444)
- validateConstraintSet (method on *Mapper) (line 448)
- solverContext (method on *Mapper) (line 468)
- solver (method on *Mapper) (line 496)
- toSolverConstraintSet (method on *Mapper) (line 504)
- shiftWhere (method on *Mapper) (line 531)
- blockKey (line 608)
- formatBlockRooms (line 617)
- cloneBlock (line 626)
- printRejection (method on *Mapper) (line 634)
- destinationCandidateBlocks (method on *Mapper) (line 654)
- moveDestinationWithCandidates (method on *Mapper) (line 761)
- validateBlockMove (method on *Mapper) (line 854)
- moveBlock (method on *Mapper) (line 923)
- captureSnapshot (method on *Mapper) (line 940)
- restoreSnapshot (method on *Mapper) (line 965)
- stateSignature (method on *Mapper) (line 982)
- holeOpenNow (method on *Mapper) (line 1000)
- smallestBlocks (line 1017)
- planningBlocks (line 1033)
- blockHasID (line 1080)
- plannerDeltas (line 1085)
- planMakeRoomMultiStepDepth (method on *Mapper) (line 1112)
- planMakeRoomMultiStep (method on *Mapper) (line 1259)
- validateHoleOpens (method on *Mapper) (line 1308)
- makeRoom (method on *Mapper) (line 1347)
- rebuildDiscoveredLayout (method on *Mapper) (line 1483)
- Enter (method on *Mapper) (line 1516)
- enterIncremental (method on *Mapper) (line 1534)
- PrintGrid10x10 (method on *Mapper) (line 1647)
- PrintGrid10x10Discovered (method on *Mapper) (line 1651)
- PrintRooms (method on *Mapper) (line 1713)
- PrintRoomsDiscovered (method on *Mapper) (line 1717)
- ExitsFrom (method on *World) (line 1772)
- Neighbors (method on *World) (line 1784)
- HasRoom (method on *World) (line 1797)
- NewDiscoveryState (line 1809)
- Discover (method on *DiscoveryState) (line 1815)
- IsDiscovered (method on *DiscoveryState) (line 1819)
- visibleExits (line 1827)
- discoveredRoomIDs (line 1837)
- LoadWorld (line 1846)
- ensureRoom (method on *World) (line 1883)
- addExit (method on *World) (line 1892)
- parseRoomFileIntoWorld (line 1900)
- Run (line 1981)

Variables:
- uiOut (line 180)

## `internal/wcs/win32/proc_windows.go`

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

## `internal/wcs/win32/shell_windows.go`

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
