/*

Copyright 2020 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.

*/
package common

func Max(x, y int32) int32 {
	if x >= y {
		return x
	}
	return y
}

func Min(x, y int32) int32 {
	if x <= y {
		return x
	}
	return y
}
