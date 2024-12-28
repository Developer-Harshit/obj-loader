package objloader

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func sliceToFloat32(parts []string) ([]float32, error) {
	result := []float32{}
	for _, v := range parts {
		num, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return result, err
		}
		result = append(result, float32(num))
	}
	return result, nil
}
func sliceToInt(parts []string) ([]int, error) {
	result := []int{}
	for _, v := range parts {
		num, err := strconv.Atoi(v)
		if err != nil {
			return result, err
		}
		result = append(result, num)
	}
	return result, nil
}

/*
IT TRIES TO FOLLOW THE FOLLOWING SPECS
https://www.martinreddy.net/gfx/3d/OBJ.spec
*/

func getFaceType(p string) int {
	v := strings.Split(p, "/")
	if len(v) == 0 {
		return -1
	} // invalid type
	if len(v) == 1 {
		return 0
	} // f v
	if len(v) == 2 {
		return 1
	} // f v/vt
	if len(v) == 3 {
		if v[1] != "" {
			return 2 // f v/vt/vn
		} else {
			return 3 // f v//vn
		}
	}
	return -1
}
func getFaceData(p string, vtype int) []int {
	result := []int{}
	reader := strings.NewReader(p)
	switch vtype {
	case 0: // f v
		var a int
		fmt.Fscanf(reader, "%d", &a)
		result = append(result, a)
	case 1: // f v/vt
		var a, b int
		fmt.Fscanf(reader, "%d/%d", &a, &b)
		result = append(result, []int{a, b}...)
	case 2: // f v/vt/vn
		var a, b, c int
		fmt.Fscanf(reader, "%d/%d/%d", &a, &b, &c)
		result = append(result, []int{a, b, c}...)
	case 3: // f v//vn
		var a, b int
		fmt.Fscanf(reader, "%d//%d", &a, &b)
		result = append(result, []int{a, b}...)
	default:
		fmt.Println(errorInvalidFormat("Invalid Vtype", vtype))
	}
	return result
}
func errorInvalidFormat(msg string, p any) error {
	return fmt.Errorf("INVALID FORMAT: %s _> %v", msg, p)
}
func errorNotSupported(msg string, p any) error {
	return fmt.Errorf("NOT SUPPORTED: %s _> %v", msg, p)
}

type Vec3 = [3]float32
type Vec2 = [2]float32

type Mesh struct {
	Vertices     []float32
	Vtype        int
	MaterialName string
}

type Options struct {
	NeedNormals bool
}

type Material struct {
	Name  string
	Ns    float32
	Ni    float32
	D     float32
	Illum int32
	Ka    Vec3
	Kd    Vec3
	Ks    Vec3
	Ke    Vec3
	Ki    Vec3
}

// It assumes mtl file have absolute path or path relative to current dir
func LoadMtl(filePath string, materials *MaterialMap) error {
	var m *Material
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		rawText := scanner.Text()
		parts := strings.Fields(rawText)
		if len(parts) == 0 {
			continue
		}
		key := parts[0]
		parts = parts[1:]
		switch key {
		case "newmtl":
			m = &Material{Name: parts[0]}
			(*materials)[parts[0]] = m
		case "Ns":
			if len(parts) < 1 {
				return errorInvalidFormat("'Ns' must have 1 element", parts)
			}
			num, err := strconv.ParseFloat(parts[0], 32)
			if err != nil {
				return err
			}
			m.Ns = float32(num)
		case "Ni":
			if len(parts) < 1 {
				return errorInvalidFormat("'Ni' must have 1 element", parts)
			}
			num, err := strconv.ParseFloat(parts[0], 32)
			if err != nil {
				return err
			}
			m.Ni = float32(num)
		case "d":
			if len(parts) < 1 {
				return errorInvalidFormat("'d' must have 1 element", parts)
			}
			num, err := strconv.ParseFloat(parts[0], 32)
			if err != nil {
				return err
			}
			m.D = float32(num)
		case "illum":
			if len(parts) < 1 {
				return errorInvalidFormat("'illum' must have 1 element", parts)
			}
			num, err := strconv.Atoi(parts[0])
			if err != nil {
				return err
			}
			m.Illum = int32(num)
		case "Ka":
			if len(parts) < 3 {
				return errorInvalidFormat("'Ka' must have 3 element", parts)
			}
			nums, err := sliceToFloat32(parts)
			if err != nil {
				return err
			}
			m.Ka = Vec3{nums[0], nums[1], nums[2]}
		case "Kd":
			if len(parts) < 3 {
				return errorInvalidFormat("'Kd' must have 3 element", parts)
			}
			nums, err := sliceToFloat32(parts)
			if err != nil {
				return err
			}
			m.Kd = Vec3{nums[0], nums[1], nums[2]}

		case "Ke":
			if len(parts) < 3 {
				return errorInvalidFormat("'Ke' must have 3 element", parts)
			}
			nums, err := sliceToFloat32(parts)
			if err != nil {
				return err
			}
			m.Ke = Vec3{nums[0], nums[1], nums[2]}

		case "Ki":
			if len(parts) < 3 {
				return errorInvalidFormat("'Ki' must have 3 element", parts)
			}
			nums, err := sliceToFloat32(parts)
			if err != nil {
				return err
			}
			m.Ki = Vec3{nums[0], nums[1], nums[2]}

		case "Ks":
			if len(parts) < 3 {
				return errorInvalidFormat("'Ks' must have 3 element", parts)
			}
			nums, err := sliceToFloat32(parts)
			if err != nil {
				return err
			}
			m.Ks = Vec3{nums[0], nums[1], nums[2]}

		}
	}
	return nil
}

type MaterialMap = map[string]*Material
type MeshMap = map[string]*Mesh

func LoadObj(filePath string, meshes *MeshMap, materials *MaterialMap, options Options) error {

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)

	positions := []Vec3{}
	normals := []Vec3{}
	uvs := []Vec2{}

	lineCounter := 0
	facesCounter := 0

	currentMaterial := "default"
	_nsMap := map[string]bool{}

	for scanner.Scan() {
		lineCounter++
		rawText := scanner.Text()
		parts := strings.Fields(rawText)
		if len(parts) == 0 {
			continue
		}
		key := parts[0]
		parts = parts[1:]
		switch key {
		case "#": // USED FOR COMMENTS

		// VERTEX DATA _> (v) (vn) (vt) (vp) (cstype) (deg) (bmat) (step)
		// VERTEX POSITIONS [x,y,z,(w)] Required
		case "v":
			if len(parts) > 3 {
				return errorInvalidFormat("Vertex Position can only have 3 elements", parts)
			}
			vert, err := sliceToFloat32(parts[:3])
			if err != nil {
				return err
			}
			positions = append(positions, Vec3{vert[0], vert[1], vert[2]})

		// SURFACE NORMALS [i,j,k] Optional / Can be recalculated using (v)
		case "vn":
			if len(parts) > 3 {
				return errorInvalidFormat("Surface Normal can only have 3 elements", parts)
			}
			norm, err := sliceToFloat32(parts[:3])
			if err != nil {
				return err
			}
			normals = append(normals, Vec3{norm[0], norm[1], norm[2]})

		// TEXTURE COORDINATES [u,v] Optional
		case "vt":
			if len(parts) > 2 {
				return errorInvalidFormat("Texture coords only have 2 elements", parts)
			}
			uv, err := sliceToFloat32(parts[:2])
			if err != nil {
				return err
			}
			uvs = append(uvs, Vec2{uv[0], uv[1]})

		// ELEMENTS _> (p) (l) (f) (curv) (curv2) (surf)
		// FACES
		case "f":

			// TODO: SUPPORT NEGATIVE INDICES
			// TODO: SUPPORT MORE THAN 3 FACES (TRIANGLE_FAN)

			// ERROR CHECKING
			vtype := getFaceType(strings.Trim(parts[0], "/"))
			if len(parts) < 3 {
				return errorInvalidFormat("Faces cannot have length less than 3", parts)
			}
			if len(parts) != 3 {
				return errorNotSupported("Only Supports faces with length 3", parts)
			}
			if vtype == -1 {
				return errorInvalidFormat("Undefined face format", parts)
			}

			faces := [][]int{}
			for _, p := range parts {
				p = strings.Trim(p, "/") // Trim leading and trailing '/'
				if vtype != getFaceType(p) {
					return errorInvalidFormat("Faces cannot have inconsistent format", parts)
				}
				faceData := getFaceData(p, vtype)
				faces = append(faces, faceData)
			}

			if len(faces)%3 != 0 {
				return errorInvalidFormat("Undefined Behavior", parts)
			}

			vertices := []float32{}
			uvIdx := -1
			nIdx := -1
			if vtype == 2 {
				nIdx = 2
			} else if vtype == 3 {
				nIdx = 1
			}
			if vtype == 2 || vtype == 1 {
				uvIdx = 1
			}

			for i := 0; i < len(faces); i += 3 {
				f1 := faces[i]
				f2 := faces[i+1]
				f3 := faces[i+2]
				// POSITIONS
				va := positions[f1[0]-1]
				vb := positions[f2[0]-1]
				vc := positions[f3[0]-1]
				// NORMALS
				var na, nb, nc Vec3
				if nIdx != -1 {
					na = normals[f1[nIdx]-1]
					nb = normals[f2[nIdx]-1]
					nc = normals[f3[nIdx]-1]
				} else if options.NeedNormals {
					u := Vec3{vb[0] - va[0], vb[1] - va[1], vb[2] - va[2]}
					v := Vec3{vc[0] - va[0], vc[1] - va[1], vc[2] - va[2]}
					n := Vec3{
						u[1]*v[2] - u[2]*v[1],
						u[2]*v[0] - u[0]*v[2],
						u[0]*v[1] - u[1]*v[0],
					}
					na = n
					nb = n
					nc = n
				}
				// UVS
				var uva, uvb, uvc Vec2
				if uvIdx != -1 {
					uva = uvs[f1[uvIdx]-1]
					uvb = uvs[f2[uvIdx]-1]
					uvc = uvs[f3[uvIdx]-1]
				}

				// v1
				vertices = append(vertices, va[:]...)
				if uvIdx != -1 {
					vertices = append(vertices, uva[:]...)
				}
				if options.NeedNormals {
					vertices = append(vertices, na[:]...)
				}
				// v2
				vertices = append(vertices, vb[:]...)
				if uvIdx != -1 {
					vertices = append(vertices, uvb[:]...)
				}
				if options.NeedNormals {
					vertices = append(vertices, nb[:]...)
				}
				// v3
				vertices = append(vertices, vc[:]...)
				if uvIdx != -1 {
					vertices = append(vertices, uvc[:]...)
				}
				if options.NeedNormals {
					vertices = append(vertices, nc[:]...)
				}
			}
			if len(vertices) != 0 {
				meshID := fmt.Sprintf("mesh-vtype%d-material%s", vtype, currentMaterial)
				_, ok := (*meshes)[meshID]
				if !ok {
					(*meshes)[meshID] = &Mesh{Vertices: []float32{}, Vtype: vtype, MaterialName: currentMaterial}
				}
				mesh := (*meshes)[meshID]
				mesh.Vertices = append(mesh.Vertices, vertices...)
			}
			facesCounter++

		// RENDER ATTRIBUTES (usemtl) (mtllib)
		// USE MATERIAL usemtl (material_name)
		case "usemtl":
			// TODO: TURN ON MATERIALS

			if len(parts) != 1 || parts[0] == "" {
				currentMaterial = "default"
			} else {
				currentMaterial = parts[0]
			}

		// MATERIAL LIBRARY mtllib filepath1 filepath2 ....
		case "mtllib":
			// TODO: LOAD MATERIALS
			for _, filePath := range parts {
				err := LoadMtl(filePath, materials)
				if err != nil {
					return err
				}
			}

		default:
			_, ok := _nsMap[key]
			if ok {
				continue
			}
			fmt.Println(errorNotSupported("The given keyword is not supported", key))
			_nsMap[key] = true
		}
	}

	err = scanner.Err()
	if err != nil {
		return err
	}

	return nil
}
