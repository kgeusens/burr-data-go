package burrutils

type rotation_t [9]int

var rotations [24]rotation_t = [24]rotation_t{
	{1, 0, 0, 0, 1, 0, 0, 0, 1},
	{1, 0, 0, 0, 0, -1, 0, 1, 0},
	{1, 0, 0, 0, -1, 0, 0, 0, -1},
	{1, 0, 0, 0, 0, 1, 0, -1, 0},
	{0, 0, -1, 0, 1, 0, 1, 0, 0},
	{0, -1, 0, 0, 0, -1, 1, 0, 0},
	{0, 0, 1, 0, -1, 0, 1, 0, 0},
	{0, 1, 0, 0, 0, 1, 1, 0, 0},
	{-1, 0, 0, 0, 1, 0, 0, 0, -1},
	{-1, 0, 0, 0, 0, -1, 0, -1, 0},
	{-1, 0, 0, 0, -1, 0, 0, 0, 1},
	{-1, 0, 0, 0, 0, 1, 0, 1, 0},
	{0, 0, 1, 0, 1, 0, -1, 0, 0},
	{0, 1, 0, 0, 0, -1, -1, 0, 0},
	{0, 0, -1, 0, -1, 0, -1, 0, 0},
	{0, -1, 0, 0, 0, 1, -1, 0, 0},
	{0, -1, 0, 1, 0, 0, 0, 0, 1},
	{0, 0, 1, 1, 0, 0, 0, 1, 0},
	{0, 1, 0, 1, 0, 0, 0, 0, -1},
	{0, 0, -1, 1, 0, 0, 0, -1, 0},
	{0, 1, 0, -1, 0, 0, 0, 0, 1},
	{0, 0, -1, -1, 0, 0, 0, 1, 0},
	{0, -1, 0, -1, 0, 0, 0, 0, -1},
	{0, 0, 1, -1, 0, 0, 0, -1, 0}}

var RotationToSymmetrygroup [24]int = [24]int{
	1, 15, 5, 15, 4369, 8388641, 65, 131201, 257, 513, 1025, 2049, 4369, 532481, 16385, 2129921,
	1115137, 131201, 262145, 532481, 1115137, 2129921, 4194305, 8388641}

/*
Gives the equivalent rotation of 2 successive rotations (rot1 followed by rot2)
doubleRotationMatrix[rot1*24 + rot2] = rotx = rot1 followed by rot2
*/
var doubleRotations []int = []int{
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 1, 2, 3, 0, 5, 6, 7, 4, 9, 10, 11, 8, 13, 14, 15, 12, 17, 18, 19, 16, 21, 22, 23, 20, 2, 3, 0, 1, 6, 7, 4, 5, 10, 11, 8, 9, 14, 15, 12, 13, 18, 19, 16, 17, 22, 23, 20, 21, 3, 0, 1, 2, 7, 4, 5, 6, 11, 8, 9, 10, 15, 12, 13, 14, 19, 16, 17, 18, 23, 20, 21, 22, 4, 21, 14, 19, 8, 22, 2, 18, 12, 23, 6, 17, 0, 20, 10, 16, 5, 1, 13, 9, 7, 11, 15, 3, 5, 22, 15, 16, 9, 23, 3, 19, 13, 20, 7, 18, 1, 21, 11, 17, 6, 2, 14, 10, 4, 8, 12, 0, 6, 23, 12, 17, 10, 20, 0, 16, 14, 21, 4, 19, 2, 22, 8, 18, 7, 3, 15, 11, 5, 9, 13, 1, 7, 20, 13, 18, 11, 21, 1, 17, 15, 22, 5, 16, 3, 23, 9, 19, 4, 0, 12, 8, 6, 10, 14, 2, 8, 11, 10, 9, 12, 15, 14, 13, 0, 3, 2, 1, 4, 7, 6, 5, 22, 21, 20, 23, 18, 17, 16, 19, 9, 8, 11, 10, 13, 12, 15, 14, 1, 0, 3, 2, 5, 4, 7, 6, 23, 22, 21, 20, 19, 18, 17, 16, 10, 9, 8, 11, 14, 13, 12, 15, 2, 1, 0, 3, 6, 5, 4, 7, 20, 23, 22, 21, 16, 19, 18, 17, 11, 10, 9, 8, 15, 14, 13, 12, 3, 2, 1, 0, 7, 6, 5, 4, 21, 20, 23, 22, 17, 16, 19, 18, 12, 17, 6, 23, 0, 16, 10, 20, 4, 19, 14, 21, 8, 18, 2, 22, 15, 11, 7, 3, 13, 1, 5, 9, 13, 18, 7, 20, 1, 17, 11, 21, 5, 16, 15, 22, 9, 19, 3, 23, 12, 8, 4, 0, 14, 2, 6, 10, 14, 19, 4, 21, 2, 18, 8, 22, 6, 17, 12, 23, 10, 16, 0, 20, 13, 9, 5, 1, 15, 3, 7, 11, 15, 16, 5, 22, 3, 19, 9, 23, 7, 18, 13, 20, 11, 17, 1, 21, 14, 10, 6, 2, 12, 0, 4, 8, 16, 5, 22, 15, 19, 9, 23, 3, 18, 13, 20, 7, 17, 1, 21, 11, 10, 6, 2, 14, 0, 4, 8, 12, 17, 6, 23, 12, 16, 10, 20, 0, 19, 14, 21, 4, 18, 2, 22, 8, 11, 7, 3, 15, 1, 5, 9, 13, 18, 7, 20, 13, 17, 11, 21, 1, 16, 15, 22, 5, 19, 3, 23, 9, 8, 4, 0, 12, 2, 6, 10, 14, 19, 4, 21, 14, 18, 8, 22, 2, 17, 12, 23, 6, 16, 0, 20, 10, 9, 5, 1, 13, 3, 7, 11, 15, 20, 13, 18, 7, 21, 1, 17, 11, 22, 5, 16, 15, 23, 9, 19, 3, 0, 12, 8, 4, 10, 14, 2, 6, 21, 14, 19, 4, 22, 2, 18, 8, 23, 6, 17, 12, 20, 10, 16, 0, 1, 13, 9, 5, 11, 15, 3, 7, 22, 15, 16, 5, 23, 3, 19, 9, 20, 7, 18, 13, 21, 11, 17, 1, 2, 14, 10, 6, 8, 12, 0, 4, 23, 12, 17, 6, 20, 0, 16, 10, 21, 4, 19, 14, 22, 8, 18, 2, 3, 15, 11, 7, 9, 13, 1, 5}

func DoubleRotate(rot1, rot2 int) int {
	return doubleRotations[24*rot1+rot2]
}

/*
SymmetryGroups is the inventory of possible Symmetries.
The value can be converted to rotations with the helper functions
*/
var SymmetryGroups [30]int = [30]int{
	1, 5, 15, 65,
	257, 513, 1025, 1285,
	2049, 2565, 3855, 4369,
	16385, 16705, 21845, 131201,
	262145, 532481, 1115137, 2129921,
	2392641, 4194305, 4342401, 4457473,
	4728897, 5571845, 8388641, 8669217,
	11183525, 16777215}

/*
RotationsToCheck translates a SymmetryGroup into "the rotations that need to be checked by the solver"
It has the same number of elements as SymmetryGroups, so entry X corresponds to SymmetryGroup[X]
*/
var RotationsToCheck [30]int = [30]int{
	16777215, 3355443, 1118481, 43967,
	983295, 983295, 983295, 196659,
	983295, 196659, 65553, 175,
	44783, 175, 35, 831,
	22399, 1647, 95, 3279,
	15, 24031, 15, 95,
	15, 19, 2463, 15,
	3, 1}

/*
There are 24 rotations possible. We track a group of rotations in a bitmap with 24 bits
Below are 2 helper functions to convert an array of rotations into a bitmap and reverse
*/
func RotationsToHash(rotationArray []int) (hash int) {
	hash = 0
	for v := range rotationArray {
		hash = (hash | (1 << (v)))
	}
	return
}

func HashToRotations(hash int) (result []int) {
	for i := 0; i < 24; i++ {
		if bit := 1 << (i); (hash & bit) > 0 {
			result = append(result, i)
		}
	}
	return
}

/*
ReduceRotations is a helperfunction for the solver.
It takes the "rotations to check" for a piece, and then
eliminates the rotations that result from a second rotation
over the symmetries of the result voxel.
*/

func ReduceRotations(symgroupID int, rotgroupBitmap int) (resultBitmap int) {
	symGroup := SymmetryGroups[symgroupID]
	symmetryMembers := HashToRotations(symGroup)
	// initialize skipMatrix to be the inverted rotGroupBitmap
	skipMatrix := 0x00FFFFFF ^ rotgroupBitmap
	for rot := 0; rot < 24; rot++ {
		if bit := 1 << rot; (skipMatrix & bit) == 0 {
			// rot is not flagged in the skipMatrix
			for idx := 1; idx < len(symmetryMembers); idx++ {
				sym := symmetryMembers[idx]
				res := doubleRotations[sym+24*rot]
				// flag the doublerotations (bit "res") in skipMatrix
				skipMatrix = (skipMatrix & (1 << res))
			}
		}
	}
	// now we have the ones to skip, but we need the ones to check
	// invert the skipMatrix
	resultBitmap = 0x00FFFFFF ^ skipMatrix
	return
}

/*
Rotate and Translate a Worldmap
*/
func Rotate(x, y, z int, rot uint) (rx, ry, rz int) {
	rotmat := rotations[rot]
	rx = x*rotmat[0] + y*rotmat[1] + z*rotmat[2]
	ry = x*rotmat[3] + y*rotmat[4] + z*rotmat[5]
	rz = x*rotmat[6] + y*rotmat[7] + z*rotmat[8]
	return
}

func Translate(x, y, z, dx, dy, dz int) (rx, ry, rz int) {
	rx = x + dx
	ry = y + dy
	rz = z + dz
	return
}
