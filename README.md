# Experiment the x/image/vector's invalid memory accesses.

Research about unexpected panics and weird drawings of `x/image/vector`'s `Draw` methods.

## Table of content

- [Test Cases](#test-cases)
  1. [Common](#1-common)
  2. [Negative](#2-negative)
  3. [NegativeY](#3-negativey)
  4. [NegativeX](#4-negativex)
  5. [Overflow](#5-overflow)
  6. [OverflowY](#6-overflowy)
  7. [OverflowX](#7-overflowx)
- [When these happens](#when-these-happens)
- [Why these happens](#why-these-happens)
- [Possible fix](#possible-fix)

## Test cases:

### 1. Common:

Nothing special, just drawing a 25x25px vector at the 0x0 position of a 50x50px image.

| Image Type | Normal                                           | RGBA Patched                                      |
| ---------- | ------------------------------------------------ | ------------------------------------------------- |
| NRGBA      | ![NRGBA Basic](./assets/Common-NRGBA-normal.png) | ![NRGBA Basic](./assets/Common-NRGBA-patched.png) |
| RGBA       | ![RGBA Basic](./assets/Common-RGBA-normal.png)   | ![RGBA Basic](./assets/Common-RGBA-patched.png)   |
| Alpha      | ![Alpha Basic](./assets/Common-Alpha-normal.png) | ![Alpha Basic](./assets/Common-Alpha-patched.png) |

### 2. Negative:

Drawing a 25x25px vector at a lower position (e.g., -10x-10) of a 50x50px image's bounds.

| Image Type | Normal                                             | RGBA Patched                                        |
| ---------- | -------------------------------------------------- | --------------------------------------------------- |
| NRGBA      | ![NRGBA Basic](./assets/Negative-NRGBA-normal.png) | ![NRGBA Basic](./assets/Negative-NRGBA-patched.png) |
| RGBA       | PANIC                                              | ![RGBA Basic](./assets/Negative-RGBA-patched.png)   |
| Alpha      | PANIC                                              | Not Patched                                         |

### 3. NegativeY:

Drawing a 25x25px vector at a lower Y position (e.g., 0x-10) of a 50x50px image's bounds.

| Image Type | Normal                                              | RGBA Patched                                         |
| ---------- | --------------------------------------------------- | ---------------------------------------------------- |
| NRGBA      | ![NRGBA Basic](./assets/NegativeY-NRGBA-normal.png) | ![NRGBA Basic](./assets/NegativeY-NRGBA-patched.png) |
| RGBA       | PANIC                                               | ![RGBA Basic](./assets/NegativeY-RGBA-patched.png)   |
| Alpha      | PANIC                                               | NOT Patched                                          |

### 4. NegativeX:

Drawing a 25x25px vector at a lower X position with a higher Y (e.g., -10x2, to avoid crashes) of a 50x50px image's bounds.

| Image Type | Normal                                              | RGBA Patched                                         |
| ---------- | --------------------------------------------------- | ---------------------------------------------------- |
| NRGBA      | ![NRGBA Basic](./assets/NegativeX-NRGBA-normal.png) | ![NRGBA Basic](./assets/NegativeX-NRGBA-patched.png) |
| RGBA       | ![RGBA Basic](./assets/NegativeX-RGBA-normal.png)   | ![RGBA Basic](./assets/NegativeX-RGBA-patched.png)   |
| Alpha      | ![Alpha Basic](./assets/NegativeX-Alpha-normal.png) | ![Alpha Basic](./assets/NegativeX-Alpha-patched.png) |

### 5. Overflow:

Drawing a 25x25px vector at a higher position (eg: 35x35) than a 50x50px image's bounds.

| Image Type | Normal                                             | RGBA Patched                                        |
| ---------- | -------------------------------------------------- | --------------------------------------------------- |
| NRGBA      | ![NRGBA Basic](./assets/Overflow-NRGBA-normal.png) | ![NRGBA Basic](./assets/Overflow-NRGBA-patched.png) |
| RGBA       | PANIC                                              | ![RGBA Basic](./assets/Overflow-RGBA-patched.png)   |
| Alpha      | PANIC                                              | Not Patched                                         |

### 6. OverflowY:

Drawing a 25x25px vector at a higher Y position (eg: 0x30) than a 50x50px image's bounds.

| Image Type | Normal                                              | RGBA Patched                                         |
| ---------- | --------------------------------------------------- | ---------------------------------------------------- |
| NRGBA      | ![NRGBA Basic](./assets/OverflowY-NRGBA-normal.png) | ![NRGBA Basic](./assets/OverflowY-NRGBA-patched.png) |
| RGBA       | PANIC                                               | ![RGBA Basic](./assets/OverflowY-RGBA-patched.png)   |
| Alpha      | PANIC                                               | Not Patched                                          |

### 7. OverflowX:

Drawing a 25x25px vector at a higher X position (eg: 35x15) than a 50x50px image's bounds.

| Image Type | Normal                                              | RGBA Patched                                         |
| ---------- | --------------------------------------------------- | ---------------------------------------------------- |
| NRGBA      | ![NRGBA Basic](./assets/OverflowX-NRGBA-normal.png) | ![NRGBA Basic](./assets/OverflowX-NRGBA-patched.png) |
| RGBA       | ![RGBA Basic](./assets/OverflowX-RGBA-normal.png)   | ![RGBA Basic](./assets/OverflowX-RGBA-patched.png)   |
| Alpha      | ![Alpha Basic](./assets/OverflowX-Alpha-normal.png) | ![Alpha Basic](./assets/OverflowX-Alpha-patched.png) |

## When these happens?

These issues will only happen if your src image is an `image.Uniform` and your dst image is one of `image.RGBA` or `image.RGBA`.

They will still draw normally if your `r`'s (second Draw method's parameter) `Min` and `Max` are within the dst's range. But things start to happen if `r` isn't  in the `dst`'s range, as you can see in the [test cases](#test-cases).

## Why these happens?

Both `image.RGBA` as `image.Alpha` have their own specific implementations for `image.Uniform` images as source. You can find them [here](https://cs.opensource.google/go/x/image/+/master:vector/vector.go;l=272;drc=cff245a6509b8c4de022d0d5b9037c503c5989d6).

I've done all of my researches over `image.RGBA`, so here's a detailed description. (most of the thing are same for `image.Alpha` too)

When we call the `Draw` method with an `image.RGBA` as dst and an `image.Uniform` as src, if we assume your `DrawOp` is `draw.Over`, it will draw the vector with the `rasterizeDstRGBASrcUniformOpOver` method. (which is usually `rasterizeOpOver` for other image types.) And this method does not depend on what `image.RGBA` actually provides; it will access and write the pixels manually by itself. Here's its code at the time of writing this:

```go
func (z *Rasterizer) rasterizeDstRGBASrcUniformOpOver(dst *image.RGBA, r image.Rectangle, sr, sg, sb, sa uint32) {
	z.accumulateMask()
	pix := dst.Pix[dst.PixOffset(r.Min.X, r.Min.Y):]
	for y, y1 := 0, r.Max.Y-r.Min.Y; y < y1; y++ {
		for x, x1 := 0, r.Max.X-r.Min.X; x < x1; x++ {
			ma := z.bufU32[y*z.size.X+x]

			// This formula is like rasterizeOpOver's, simplified for the
			// concrete dst type and uniform src assumption.
			a := 0xffff - (sa * ma / 0xffff)
			i := y*dst.Stride + 4*x
			pix[i+0] = uint8(((uint32(pix[i+0])*0x101*a + sr*ma) / 0xffff) >> 8)
			pix[i+1] = uint8(((uint32(pix[i+1])*0x101*a + sg*ma) / 0xffff) >> 8)
			pix[i+2] = uint8(((uint32(pix[i+2])*0x101*a + sb*ma) / 0xffff) >> 8)
			pix[i+3] = uint8(((uint32(pix[i+3])*0x101*a + sa*ma) / 0xffff) >> 8)
		}
	}
}
```

A few issues could happen with this in different cases.

1. When you pass a smaller `r.Min` than `dst`'s Min.

   At this moment, it tried to access the Pix array with a negative number gathered from `dst.PixOffset`, and as you already know, you can't access an array with a negative index.

   Test cases: [Negative](#2-negative), [NegativeY](#3-negativey)

2. When you pass smaller `r.Min.X` or bigger `r.Min.X` than `dst`'s bounds.

   As you can see, it just loops through the `r`'s (vector's) bounds and writes them directly on the `dst` image without taking care of the `dst`'s limits. This causes two things: possible crashes (min=(negative X, zero Y) or max=(higher X, highest possible Y)), or drawing the overflowed parts of the vector on the other side of the image a single pixel higher or smaller (depending on your rect's min and max.).

   Test cases: [NegativeX](#4-negativex), [OverflowX](#7-overflowx)

3. When you pass a higher `r.Max` than `dst`'s Max.

   Same as the first case, but happens when we're drawing the pixels. It happens when there's an overflow of the vector, and it _should be_ skipped. But as vector doesn't take care of the `dst`'s bounds, it will try to access or write an item out of `dst.Pix`'s range. (the last 4 lines of the code above.)

   Test cases: [Overflow](#5-overflow), [OverflowY](#6-overflowy)

## Possible fix

TODO
