# Go Source Index

Generated from current `.go` files. Use this as a quick technical map for chat/session continuity.

## Go File Tree

- `internal\wcs\win32\proc_windows.go`
- `internal\wcs\win32\shell_windows.go`
- `main.go`
- `navigation_session.go`
- `navigation_session_test.go`
- `solver\solver.go`

## `internal\wcs\win32\proc_windows.go`

Types:
- (none)

Functions:
- (none)

## `internal\wcs\win32\shell_windows.go`

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

## `main.go`

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
- node (line 1113)
- Topology (line 1761)
- World (line 1767)
- DiscoveryState (line 1805)

Functions:
- relationForKey (line 68)
- BuildConstraintSet (method on m *Mapper) (line 86)
- Error (method on e *lockedAdjViolationError) (line 151)
- Error (method on e *collisionError) (line 163)
- uiPrint (line 182)
- uiPrintf (line 186)
- uiPrintln (line 190)
- setupLogging (line 194)
- NewMapper (line 209)
- BindTopology (method on m *Mapper) (line 220)
- SetDebugWriter (method on m *Mapper) (line 224)
- SetSolverProvider (method on m *Mapper) (line 232)
- debugln (method on m *Mapper) (line 240)
- debugf (method on m *Mapper) (line 244)
- colName (line 248)
- cellLabel (line 250)
- normalizeDirName (line 263)
- dirDelta (line 289)
- getRoom (method on m *Mapper) (line 314)
- clearOcc (method on m *Mapper) (line 323)
- setOcc (method on m *Mapper) (line 331)
- edgeAlignedAndOrdered (line 352)
- roomBetweenAxis (line 375)
- noRoomBetweenAxis (method on m *Mapper) (line 393)
- refreshLockedAdjacencies (method on m *Mapper) (line 413)
- validateLockedAdjacencies (method on m *Mapper) (line 444)
- validateConstraintSet (method on m *Mapper) (line 448)
- solverContext (method on m *Mapper) (line 468)
- solver (method on m *Mapper) (line 496)
- toSolverConstraintSet (method on m *Mapper) (line 504)
- shiftWhere (method on m *Mapper) (line 531)
- blockKey (line 608)
- formatBlockRooms (line 617)
- cloneBlock (line 626)
- printRejection (method on m *Mapper) (line 634)
- destinationCandidateBlocks (method on m *Mapper) (line 654)
- moveDestinationWithCandidates (method on m *Mapper) (line 761)
- validateBlockMove (method on m *Mapper) (line 854)
- moveBlock (method on m *Mapper) (line 923)
- captureSnapshot (method on m *Mapper) (line 940)
- restoreSnapshot (method on m *Mapper) (line 965)
- stateSignature (method on m *Mapper) (line 982)
- holeOpenNow (method on m *Mapper) (line 1000)
- smallestBlocks (line 1017)
- planningBlocks (line 1033)
- blockHasID (line 1080)
- plannerDeltas (line 1085)
- planMakeRoomMultiStepDepth (method on m *Mapper) (line 1112)
- planMakeRoomMultiStep (method on m *Mapper) (line 1259)
- validateHoleOpens (method on m *Mapper) (line 1308)
- makeRoom (method on m *Mapper) (line 1347)
- rebuildDiscoveredLayout (method on m *Mapper) (line 1483)
- Enter (method on m *Mapper) (line 1516)
- enterIncremental (method on m *Mapper) (line 1534)
- PrintGrid10x10 (method on m *Mapper) (line 1647)
- PrintGrid10x10Discovered (method on m *Mapper) (line 1651)
- PrintRooms (method on m *Mapper) (line 1713)
- PrintRoomsDiscovered (method on m *Mapper) (line 1717)
- ExitsFrom (method on w *World) (line 1772)
- Neighbors (method on w *World) (line 1784)
- HasRoom (method on w *World) (line 1797)
- NewDiscoveryState (line 1809)
- Discover (method on d *DiscoveryState) (line 1815)
- IsDiscovered (method on d *DiscoveryState) (line 1819)
- visibleExits (line 1827)
- discoveredRoomIDs (line 1837)
- LoadWorld (line 1846)
- ensureRoom (method on w *World) (line 1883)
- addExit (method on w *World) (line 1892)
- parseRoomFileIntoWorld (line 1900)
- main (line 1981)

## `navigation_session.go`

Types:
- Navigator (line 6)
- NavigationSession (line 16)

Functions:
- NewNavigationSession (line 23)
- CurrentRoom (method on s *NavigationSession) (line 49)
- CurrentExits (method on s *NavigationSession) (line 53)
- Move (method on s *NavigationSession) (line 57)
- Mapper (method on s *NavigationSession) (line 70)
- Discovery (method on s *NavigationSession) (line 74)

## `navigation_session_test.go`

Types:
- (none)

Functions:
- testWorld (line 5)
- TestNavigationSessionStartAndMove (line 15)
- TestNavigationSessionNoExit (line 42)

## `solver\solver.go`

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
- edge (line 155)
- qItem (line 202)

Functions:
- Error (method on e *LockedAdjViolationError) (line 73)
- NewConstraintSolver (line 82)
- ValidateConstraintSet (method on s *ConstraintSolver) (line 92)
- ComputeRebuildResult (method on s *ConstraintSolver) (line 124)
