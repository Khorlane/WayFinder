# Go Source Index

Generated from current `.go` files using Go AST (`go/parser`, `go/ast`, `go/token`). Use this as a quick technical map for chat/session continuity.

## Go File Tree

- `docs/dev/scripts/generate_go_source_index.go`
- `main.go`
- `solver/solver.go`
- `wcs/win32/proc_windows.go`
- `wcs/win32/shell_windows.go`
- `wmr/runtime.go`
- `wne/navigation_session.go`
- `wne/navigation_session_test.go`

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

## `wmr/runtime.go`

Types:
- RoomID (line 29)
- Room (line 31)
- Mapper (line 37)
- lockedAdjKey (line 47)
- ConstraintRelation (line 53)
- ConstraintSet (line 64)
- lockedAdjViolationError (line 146)
- collisionError (line 157)
- roomSnapshot (line 169)
- mapperSnapshot (line 175)
- plannedMove (line 1011)
- Topology (line 1762)
- World (line 1768)
- DiscoveryState (line 1806)
- wneTopologyAdapter (line 1850)
- wneWorldAdapter (line 1876)
- wneMapperAdapter (line 1902)
- wneDiscoveryAdapter (line 1914)

Functions:
- relationForKey (line 69)
- BuildConstraintSet (method on *Mapper) (line 87)
- Error (method on *lockedAdjViolationError) (line 152)
- Error (method on *collisionError) (line 164)
- uiPrint (line 183)
- uiPrintf (line 187)
- uiPrintln (line 191)
- setupLogging (line 195)
- NewMapper (line 210)
- BindTopology (method on *Mapper) (line 221)
- SetDebugWriter (method on *Mapper) (line 225)
- SetSolverProvider (method on *Mapper) (line 233)
- debugln (method on *Mapper) (line 241)
- debugf (method on *Mapper) (line 245)
- colName (line 249)
- cellLabel (line 251)
- normalizeDirName (line 264)
- dirDelta (line 290)
- getRoom (method on *Mapper) (line 315)
- clearOcc (method on *Mapper) (line 324)
- setOcc (method on *Mapper) (line 332)
- edgeAlignedAndOrdered (line 353)
- roomBetweenAxis (line 376)
- noRoomBetweenAxis (method on *Mapper) (line 394)
- refreshLockedAdjacencies (method on *Mapper) (line 414)
- validateLockedAdjacencies (method on *Mapper) (line 445)
- validateConstraintSet (method on *Mapper) (line 449)
- solverContext (method on *Mapper) (line 469)
- solver (method on *Mapper) (line 497)
- toSolverConstraintSet (method on *Mapper) (line 505)
- shiftWhere (method on *Mapper) (line 532)
- blockKey (line 609)
- formatBlockRooms (line 618)
- cloneBlock (line 627)
- printRejection (method on *Mapper) (line 635)
- destinationCandidateBlocks (method on *Mapper) (line 655)
- moveDestinationWithCandidates (method on *Mapper) (line 762)
- validateBlockMove (method on *Mapper) (line 855)
- moveBlock (method on *Mapper) (line 924)
- captureSnapshot (method on *Mapper) (line 941)
- restoreSnapshot (method on *Mapper) (line 966)
- stateSignature (method on *Mapper) (line 983)
- holeOpenNow (method on *Mapper) (line 1001)
- smallestBlocks (line 1018)
- planningBlocks (line 1034)
- blockHasID (line 1081)
- plannerDeltas (line 1086)
- planMakeRoomMultiStepDepth (method on *Mapper) (line 1113)
- planMakeRoomMultiStep (method on *Mapper) (line 1260)
- validateHoleOpens (method on *Mapper) (line 1309)
- makeRoom (method on *Mapper) (line 1348)
- rebuildDiscoveredLayout (method on *Mapper) (line 1484)
- Enter (method on *Mapper) (line 1517)
- enterIncremental (method on *Mapper) (line 1535)
- PrintGrid10x10 (method on *Mapper) (line 1648)
- PrintGrid10x10Discovered (method on *Mapper) (line 1652)
- PrintRooms (method on *Mapper) (line 1714)
- PrintRoomsDiscovered (method on *Mapper) (line 1718)
- ExitsFrom (method on *World) (line 1773)
- Neighbors (method on *World) (line 1785)
- HasRoom (method on *World) (line 1798)
- NewDiscoveryState (line 1810)
- Discover (method on *DiscoveryState) (line 1816)
- IsDiscovered (method on *DiscoveryState) (line 1820)
- visibleExits (line 1828)
- discoveredRoomIDs (line 1838)
- toWNERoomID (line 1847)
- fromWNERoomID (line 1848)
- ExitsFrom (method on wneTopologyAdapter) (line 1854)
- Neighbors (method on wneTopologyAdapter) (line 1863)
- HasRoom (method on wneTopologyAdapter) (line 1872)
- ExitsFrom (method on wneWorldAdapter) (line 1880)
- Neighbors (method on wneWorldAdapter) (line 1889)
- HasRoom (method on wneWorldAdapter) (line 1898)
- BindTopology (method on wneMapperAdapter) (line 1906)
- Enter (method on wneMapperAdapter) (line 1910)
- Discover (method on wneDiscoveryAdapter) (line 1918)
- LoadWorld (line 1922)
- ensureRoom (method on *World) (line 1959)
- addExit (method on *World) (line 1968)
- parseRoomFileIntoWorld (line 1976)
- Run (line 2057)

Variables:
- uiOut (line 181)

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
