package codec

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"testing"
)

var (
	testData = []byte("=�\\i����a]��� q+#%I��*e��Ph�@����0:�����j��B��(�&�5�rMt�,쉮����db@�B����_:�N���]Q�!��,��DL8Os~�#�\u05CF�\n�D���'��喚P�̕�Z�+�� =s�$�J�WE����W|�z���T��kT�zQ�Dz�ޫ\\j�*�3�̸�hW��^}嬓$��-_W~�w�P��Ӥ�]�3v�3�,������u�B���PonY�hR&b�QN|q�6�*�I\n����\n7�\\c���p�`��f,�A[^c�KL>��Z.��/S>5�|�-���9b�v�\"�܈�����ďpثK��>�b���bX?��\u05EF���-FaV�I8)����\"�D!�U��֛J�E�S��ݴ�G&Ih�ġs��9Y�6��Nz�\u07B3�RSR�GZmܯN�ʀkC�g�b2BJyӶ+�=Zk�٥=�(��y�.g*\\�\n                                �H��})c�9̡$!��3 +�����ڭZ\n                                                       �Jń�0Oyϒli^3�D<��J����DH�i��Ye\\C�;�Sξy�φ>t>������)7V�Ω՜��>~�SB�ٵַ�\n                                                                                                                        C7Z��K�n�~(O�oҫ��{��[U��]��b�2�\n                   �]�:_:%\n�Ec�V���4�HSj=,�&���Iy�J9�7N��O=�#��u|��Ui�\\0����w�y�[B�Ԥo3���7x�{��N\nw���    ��2C�����pt��\nkk����G���4�/M���gi�eω�q�ѯS�i\tr�Tw�N�#*��7���2V�2��S\n�RSz�и$-)%��KMT���s@�-�]5lnԔ�l��H&���J�g+�eq���\tf\n                                                 �B.\n                                                    5w��$�$\"�}�<5މ\n                                                                  g3$>�O\n�'��H�Y��                                                               �ɭ��Igb,��R]\n         �������&iW��C��gM�[^Y�d��D^&�*Qh�Lt�Hp/z6�Y�E�@)H�\\�&��\nV֥�b�lһ�gj��yb�]���\n    *X\u007Fq�y�-��hx�=tζ�S��wS�c2�T�vi�i,hp��JE)O�$hA\"�؏dl��d� =I/1��4\u007F B�\tթ+&y�}d���F�B0�rQip\ny���\n�ZA���Z�JG?\n1Gx        P��\tDU�\n   �b%I\n���ˡ�\nzH뤙i�4f�#�䚪+3�3�;�!&D�:�i�̨�J���s$cD����jw���D{Θ�����_ۊ�^A�X�^m�!�П�.c����1j�wҙBh.&Un���x>:��}����7��7R��W���oG�#c��-d�\n                                                                                                                        Jv0�ƇZ��Hh�j��$�s#�/�e<J�D��Bw�g��/^c�giJO9����č��}���UIH�Os+T���;���d���<]+Te��)l<��mU��▒��␍���)�└@K�Տ±��⎽+��K�K8ӡ���8�.%�≠��⎼␊��$│<�\n���%��\n'�?����?4$I�\"�\t�'[E�␍�)I0�$±≤��4䐨,�┌������␉┌�\n��1&␋D&Y2������]0�┼�)┬/F�┬/�;>]ZN┐����VD��A┐#�ʑC�␊3��7K.E�Q�ML֔)YπJ����J␊�\U00099811␉/^�⎽�#�)␌�F▒�␤£O\n                                                                                           ┘�F⎽%��┬≥�2��␉���┴�6������7Ց@�$�^[>�����?���+�0A�ee\t.kNe�Į�mD)�fr9\u07B3�$�l�9�d/TK)f�N�̑��{F�;�:u\n                                                        �O�P�V<�x[xе���z��[��x̞�7T�7�W�Vg��>�bt�D�㡖=��I\"�~x��M�,y5`e�Z]7*�������\nXk(s��C���L�ȼ��\u07B2�U�\t;\n                         Rw�yk=e��x�\n                                    �OM7)����گ�}�U��F: P�Y��th�\"mJ��\n���\t�\n-�vS����\u007F�2q��9&�I�s�B�i\n���5�]�12��z�e�*6@��\"O�(�E\tZ�+i�d�0�]�P�U��҈;�CZ,��\n                                                       OoiR0�pRu��.9�dg:-,�S��B�0���>�3\n                                                                                       K���rV��&H@O-�șP���F\n                                                                                                           cx;\nκݖ�iw*g#X�щ�6��F�7GjV�#�4��j�G��zt��,��[&�,O9�τ̗�|��s�(P��j�*JAdt�������*��&׆��)�z�!�VK���Q�z���yS�E�\\y.:ReJГ|#B�`����5'T�,�:�9t��;c���щ��h-%��{$�U1�np���kIA�$���HM�dfJ�g�|�6�\\WH�O�晴��y��ŐS��;R]����7.T��U�^�l|�P�\n                                                                                @��F~�֧�£�⎽/�Q��\n─����\n=�/�▒��)�؆��*�0�,�,�֤��7Zo\ne%�M���K4R���v�S>�pWde���5J��y��OJ�]h��:=;�e�S�vݒL���V����֚4w�/�'3\nl$�)z��\"�kޓ�ḭ�}/,�<��B<�݂�4�a(��:Aˊ\n� ���-@�\n���{İ���l���v�^��6擕����ʼ�k.;���\\ūM<��g���+G�.��-�1H`ɛ\n                                                      �4D.1&�H�ʎ&)����+bD��L���G�e���$�1)��T�-*��YB;(v��3��Dp(t2t-���,��f�q:m:�L��څW�ot��P��㡯i��\"����Doux>E�>\ny�����\nC��+�`gK8�BY0�l��j����SEahї�\n                            ]���Ȍ*���8�o��:�.[+���[6]J�.��(�ȾT�����Ԅڢ�S��&�/��W�*q.rT@�Q�g�ɲ���\n                                                                                               8�P\\�DEvB���T�c�w��Ұ{�U�Q9F���4�����Z�H$��D�$�bE�7(Yl�=lڳS.g�ri>Vݷk[r�'0���O�,�.Fnh�q�/@�\n       KtV�{hD�t�I�h��35:T�,r���E��-(-��(6����.~W�����k�>�*^���y��5�$��V��N\n                                                                           �z��閮V[�qo\n                                                                                      �\\n;\n�Ng;Y�ѹ�=9��x��'�M���V$�����i��<�l�����^�DP�`�\"{�2������j���_���e7Q�(��ײ�\"_*c<���<��и")
)

func TestNormalizeAudio(t *testing.T) {
	hasher := sha256.New()
	NormalizeAudio(testData, 0.452)
	hasher.Write(testData)
	if result := fmt.Sprintf("%x", hasher.Sum(nil)); result != "d91e8b11a594be08b9d629cb1c4d78cb1e04cd7637768cfbe3e0f7a98bbb08dd" {
		t.Fatalf("Expected: d91e8b11a594be08b9d629cb1c4d78cb1e04cd7637768cfbe3e0f7a98bbb08dd\nGot: %x\n", result)
	}
}

func BenchmarkNormalizeAudio_Binary(b *testing.B) {
	for i := 1; i <= 12; i++ {
		b.Run(fmt.Sprintf("BenchmarkAdjuster_%dbytes", i*len(testData)), func(b2 *testing.B) {
			buf := bytes.Repeat(testData, i)
			b2.ResetTimer()
			for i2 := 0; i2 < b2.N; i2++ {
				NormalizeAudio(buf, 0.452)
			}
		})
	}
}
