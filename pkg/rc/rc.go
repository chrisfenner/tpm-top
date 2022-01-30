package rc

import (
	"fmt"
)

type rcDetails struct {
	// name is the symbol name in the TPM specification.
	name string
	// description is the description in the TPM specification.
	description string
}

type Ver1Error struct {
	raw       int
	errorCode int
}

func (v Ver1Error) Error() string {
	var name, description string
	details, ok := ver1RespCodes[v.errorCode]
	if ok {
		name = details.name
		description = details.description
	} else {
		name = "<unknown>"
		description = "Unrecognized VER1 error."
	}
	return fmt.Sprintf("(0x%x) %s: %s", v.raw, name, description)
}

// relation captures the handle, parameter, session bits for a FMT error code.
type relation int

const (
	handle relation = iota
	parameter
	session
)

func (r relation) String() string {
	switch r {
	case handle:
		return "handle"
	case parameter:
		return "parameter"
	case session:
		return "session"
	}
	return ""
}

type Fmt1Error struct {
	raw       int
	errorCode int
	rel       relation
	idx       int
}

func (f Fmt1Error) Error() string {
	var name, description string
	details, ok := fmt1RespCodes[f.errorCode]
	if ok {
		name = details.name
		description = details.description
	} else {
		name = "<unknown)"
		description = "Unrecognized FMT1 error."
	}
	if f.idx != 0 {
		return fmt.Sprintf("(0x%x) %s: %s (%s %d)", f.raw, name, description, f.rel, f.idx)
	}
	return fmt.Sprintf("(0x%x) %s: %s", f.raw, name, description)
}

type Warning struct {
	raw       int
	errorCode int
}

func (w Warning) Error() string {
	var name, description string
	details, ok := warningRespCodes[w.errorCode]
	if ok {
		name = details.name
		description = details.description
	} else {
		name = "<unknown>"
		description = "Unrecognized VER1 error."
	}
	return fmt.Sprintf("(0x%x) %s: %s", w.raw, name, description)
}

func MakeError(rc int) error {
	if rc == 0 {
		return nil
	}
	if rc < 0 {
		return fmt.Errorf("invalid TPM RC value: %d", rc)
	}
	// Check for FMT0 error:
	if (rc & 0x80) == 0 {
		// Check for TPM 2.0-defined error:
		if (rc & 0x100) != 0 {
			// Check for warning:
			if (rc & 0x800) != 0 {
				return Warning{
					raw:       rc,
					errorCode: rc & 0x7f,
				}
			}
			// Not a warning.
			return Ver1Error{
				raw:       rc,
				errorCode: rc & 0x7f,
			}
		}
		return fmt.Errorf("TPM 1.2 error: 0x%x", rc)
	}
	// FMT1 error.
	// Check if parameter-related:
	var r relation
	var i int
	if (rc & 0x40) != 0 {
		r = parameter
		i = (rc & 0xf00) >> 8
	} else {
		// Check if session-related:
		if (rc & 0x800) != 0 {
			r = session
			// Otherwise, it's handle-related.
		} else {
			r = handle
		}
		i = (rc & 0x700) >> 8
	}
	return Fmt1Error{
		raw:       rc,
		errorCode: rc & 0x3f,
		rel:       r,
		idx:       i,
	}
}

var (
	// VER1 response codes
	ver1RespCodes = map[int]rcDetails{
		0x000: rcDetails{
			"TPM_RC_INITIALIZE",
			"TPM not initialized by TPM2_Startup or already initialized commands not being accepted because of a TPM failure",
		},
		0x001: rcDetails{
			"TPM_RC_FAILURE",
			"commands not being accepted because of a TPM failure",
		},
		0x003: rcDetails{
			"TPM_RC_SEQUENCE",
			"improper use of a sequence handle",
		},
		0x00B: rcDetails{
			"TPM_RC_PRIVATE",
			"not currently used",
		},
		0x019: rcDetails{
			"TPM_RC_HMAC",
			"not currently used",
		},
		0x020: rcDetails{
			"TPM_RC_DISABLED",
			"the command is disabled",
		},
		0x021: rcDetails{
			"TPM_RC_EXCLUSIVE",
			"command failed because audit sequence required exclusivity",
		},
		0x024: rcDetails{
			"TPM_RC_AUTH_TYPE",
			"authorization handle is not correct for command",
		},
		0x025: rcDetails{
			"TPM_RC_AUTH_MISSING",
			"command requires an authorization session for handle and it is not present.",
		},
		0x026: rcDetails{
			"TPM_RC_POLICY",
			"policy failure in math operation or an invalid authPolicy value",
		},
		0x027: rcDetails{
			"TPM_RC_PCR",
			"PCR check fail",
		},
		0x028: rcDetails{
			"TPM_RC_PCR_CHANGED",
			"PCR have changed since checked.",
		},
		0x02D: rcDetails{
			"TPM_RC_UPGRADE",
			"for all commands other than TPM2_FieldUpgradeData(), this code indicates that the TPM is in field upgrade mode; for TPM2_FieldUpgradeData(), this code indicates that the TPM is not in field upgrade mode",
		},
		0x02E: rcDetails{
			"TPM_RC_TOO_MANY_CONTEXTS",
			"context ID counter is at maximum.",
		},
		0x02F: rcDetails{
			"TPM_RC_AUTH_UNAVAILABLE",
			"authValue or authPolicy is not available for selected entity.",
		},
		0x030: rcDetails{
			"TPM_RC_REBOOT",
			"a _TPM_Init and Startup(CLEAR) is required before the TPM can resume operation.",
		},
		0x031: rcDetails{
			"TPM_RC_UNBALANCED",
			"the protection algorithms (hash and symmetric) are not reasonably balanced. The digest size of the hash must be larger than the key size of the symmetric algorithm.  This may be returned by TPM2_GetTestResult() as the testResult parameter.",
		},
		0x042: rcDetails{
			"TPM_RC_COMMAND_SIZE",
			"command commandSize value is inconsistent with contents of the command buffer; either the size is not the same as the octets loaded by the hardware interface layer or the value is not large enough to hold a command header",
		},
		0x043: rcDetails{
			"TPM_RC_COMMAND_CODE",
			"command code not supported",
		},
		0x044: rcDetails{
			"TPM_RC_AUTHSIZE",
			"the value of authorizationSize is out of range or the number of octets in the Authorization Area is greater than required",
		},
		0x045: rcDetails{
			"TPM_RC_AUTH_CONTEXT",
			"use of an authorization session with a context command or another command that cannot have an authorization session.",
		},
		0x046: rcDetails{
			"TPM_RC_NV_RANGE",
			"NV offset+size is out of range.",
		},
		0x047: rcDetails{
			"TPM_RC_NV_SIZE",
			"Requested allocation size is larger than allowed.",
		},
		0x048: rcDetails{
			"TPM_RC_NV_LOCKED",
			"NV access locked.",
		},
		0x049: rcDetails{
			"TPM_RC_NV_AUTHORIZATION",
			"NV access authorization fails in command actions (this failure does not affect lockout.action)",
		},
		0x04A: rcDetails{
			"TPM_RC_NV_UNINITIALIZED",
			"an NV Index is used before being initialized or the state saved by TPM2_Shutdown(STATE) could not be restored",
		},
		0x04B: rcDetails{
			"TPM_RC_NV_SPACE",
			"insufficient space for NV allocation",
		},
		0x04C: rcDetails{
			"TPM_RC_NV_DEFINED",
			"NV Index or persistent object already defined",
		},
		0x050: rcDetails{
			"TPM_RC_BAD_CONTEXT",
			"context in TPM2_ContextLoad() is not valid",
		},
		0x051: rcDetails{
			"TPM_RC_CPHASH",
			"cpHash value already set or not correct for use",
		},
		0x052: rcDetails{
			"TPM_RC_PARENT",
			"handle for parent is not a valid parent",
		},
		0x053: rcDetails{
			"TPM_RC_NEEDS_TEST",
			"some function needs testing.",
		},
		0x054: rcDetails{
			"TPM_RC_NO_RESULT",
			"returned when an internal function cannot process a request due to an unspecified problem. This code is usually related to invalid parameters that are not properly filtered by the input unmarshaling code.",
		},
		0x055: rcDetails{
			"TPM_RC_SENSITIVE",
			"the sensitive area did not unmarshal correctly after decryption – this code is used in lieu of the other unmarshaling errors so that an attacker cannot determine where the unmarshaling error occurred",
		},
	}

	fmt1RespCodes = map[int]rcDetails{
		0x001: rcDetails{
			"TPM_RC_ASYMMETRIC",
			"asymmetric algorithm not supported or not correct",
		},
		0x002: rcDetails{
			"TPM_RC_ATTRIBUTES",
			"inconsistent attributes",
		},
		0x003: rcDetails{
			"TPM_RC_HASH",
			"hash algorithm not supported or not appropriate",
		},
		0x004: rcDetails{
			"TPM_RC_VALUE",
			"value is out of range or is not correct for the context",
		},
		0x005: rcDetails{
			"TPM_RC_HIERARCHY",
			"hierarchy is not enabled or is not correct for the use",
		},
		0x007: rcDetails{
			"TPM_RC_KEY_SIZE",
			"key size is not supported",
		},
		0x008: rcDetails{
			"TPM_RC_MGF",
			"mask generation function not supported",
		},
		0x009: rcDetails{
			"TPM_RC_MODE",
			"mode of operation not supported",
		},
		0x00A: rcDetails{
			"TPM_RC_TYPE",
			"the type of the value is not appropriate for the use",
		},
		0x00B: rcDetails{
			"TPM_RC_HANDLE",
			"the handle is not correct for the use",
		},
		0x00C: rcDetails{
			"TPM_RC_KDF",
			"unsupported key derivation function or function not appropriate for use",
		},
		0x00D: rcDetails{
			"TPM_RC_RANGE",
			"value was out of allowed range.",
		},
		0x00E: rcDetails{
			"TPM_RC_AUTH_FAIL",
			"the authorization HMAC check failed and DA counter incremented",
		},
		0x00F: rcDetails{
			"TPM_RC_NONCE",
			"invalid nonce size or nonce value mismatch",
		},
		0x010: rcDetails{
			"TPM_RC_PP",
			"authorization requires assertion of PP",
		},
		0x012: rcDetails{
			"TPM_RC_SCHEME",
			"unsupported or incompatible scheme",
		},
		0x015: rcDetails{
			"TPM_RC_SIZE",
			"structure is the wrong size",
		},
		0x016: rcDetails{
			"TPM_RC_SYMMETRIC",
			"unsupported symmetric algorithm or key size, or not appropriate for instance",
		},
		0x017: rcDetails{
			"TPM_RC_TAG",
			"incorrect structure tag",
		},
		0x018: rcDetails{
			"TPM_RC_SELECTOR",
			"union selector is incorrect",
		},
		0x01A: rcDetails{
			"TPM_RC_INSUFFICIENT",
			"the TPM was unable to unmarshal a value because there were not enough octets in the input buffer",
		},
		0x01B: rcDetails{
			"TPM_RC_SIGNATURE",
			"the signature is not valid",
		},
		0x01C: rcDetails{
			"TPM_RC_KEY",
			"key fields are not compatible with the selected use",
		},
		0x01D: rcDetails{
			"TPM_RC_POLICY_FAIL",
			"a policy check failed",
		},
		0x01F: rcDetails{
			"TPM_RC_INTEGRITY",
			"integrity check failed",
		},
		0x020: rcDetails{
			"TPM_RC_TICKET",
			"invalid ticket",
		},
		0x021: rcDetails{
			"TPM_RC_RESERVED_BITS",
			"reserved bits not set to zero as required",
		},
		0x022: rcDetails{
			"TPM_RC_BAD_AUTH",
			"authorization failure without DA implications",
		},
		0x023: rcDetails{
			"TPM_RC_EXPIRED",
			"the policy has expired",
		},
		0x024: rcDetails{
			"TPM_RC_POLICY_CC",
			"the commandCode in the policy is not the commandCode of the command or the command code in a policy command references a command that is not implemented",
		},
		0x025: rcDetails{
			"TPM_RC_BINDING",
			"public and sensitive portions of an object are not cryptographically bound",
		},
		0x026: rcDetails{
			"TPM_RC_CURVE",
			"curve not supported",
		},
		0x027: rcDetails{
			"TPM_RC_ECC_POINT",
			"point is not on the required curve.",
		},
	}

	warningRespCodes = map[int]rcDetails{
		0x001: rcDetails{
			"TPM_RC_CONTEXT_GAP",
			"gap for context ID is too large",
		},
		0x002: rcDetails{
			"TPM_RC_OBJECT_MEMORY",
			"out of memory for object contexts",
		},
		0x003: rcDetails{
			"TPM_RC_SESSION_MEMORY",
			"out of memory for session contexts",
		},
		0x004: rcDetails{
			"TPM_RC_MEMORY",
			"out of shared object/session memory or need space for internal operations",
		},
		0x005: rcDetails{
			"TPM_RC_SESSION_HANDLES",
			"out of session handles – a session must be flushed before a new session may be created out of object handles – the handle space for objects is depleted and a reboot is required",
		},
		0x006: rcDetails{
			"TPM_RC_OBJECT_HANDLES",
			"out of object handles – the handle space for objects is depleted and a reboot is required",
		},
		0x007: rcDetails{
			"TPM_RC_LOCALITY",
			"bad locality",
		},
		0x008: rcDetails{
			"TPM_RC_YIELDED",
			"the TPM has suspended operation on the command; forward progress was made and the command may be retried",
		},
		0x009: rcDetails{
			"TPM_RC_CANCELED",
			"the command was canceled",
		},
		0x00A: rcDetails{
			"TPM_RC_TESTING",
			"TPM is performing self-tests",
		},
		0x010: rcDetails{
			"TPM_RC_REFERENCE_H0",
			"the 1st handle in the handle area references a transient object or session that is not loaded",
		},
		0x011: rcDetails{
			"TPM_RC_REFERENCE_H1",
			"the 2nd handle in the handle area references a transient object or session that is not loaded",
		},
		0x012: rcDetails{
			"TPM_RC_REFERENCE_H2",
			"the 3rd handle in the handle area references a transient object or session that is not loaded",
		},
		0x013: rcDetails{
			"TPM_RC_REFERENCE_H3",
			"the 4th handle in the handle area references a transient object or session that is not loaded",
		},
		0x014: rcDetails{
			"TPM_RC_REFERENCE_H4",
			"the 5th handle in the handle area references a transient object or session that is not loaded",
		},
		0x015: rcDetails{
			"TPM_RC_REFERENCE_H5",
			"the 6th handle in the handle area references a transient object or session that is not loaded",
		},
		0x016: rcDetails{
			"TPM_RC_REFERENCE_H6",
			"the 7th handle in the handle area references a transient object or session that is not loaded",
		},
		0x018: rcDetails{
			"TPM_RC_REFERENCE_S0",
			"the 1st authorization session handle references a session that is not loaded",
		},
		0x019: rcDetails{
			"TPM_RC_REFERENCE_S1",
			"the 2nd authorization session handle references a session that is not loaded",
		},
		0x01A: rcDetails{
			"TPM_RC_REFERENCE_S2",
			"the 3rd authorization session handle references a session that is not loaded",
		},
		0x01B: rcDetails{
			"TPM_RC_REFERENCE_S3",
			"the 4th authorization session handle references a session that is not loaded",
		},
		0x01C: rcDetails{
			"TPM_RC_REFERENCE_S4",
			"the 5th session handle references a session that is not loaded",
		},
		0x01D: rcDetails{
			"TPM_RC_REFERENCE_S5",
			"the 6th session handle references a session that is not loaded",
		},
		0x01E: rcDetails{
			"TPM_RC_REFERENCE_S6",
			"the 7th authorization session handle references a session that is not loaded",
		},
		0x020: rcDetails{
			"TPM_RC_NV_RATE",
			"the TPM is rate-limiting accesses to prevent wearout of NV",
		},
		0x021: rcDetails{
			"TPM_RC_LOCKOUT",
			"authorizations for objects subject to DA protection are not allowed at this time because the TPM is in DA lockout mode",
		},
		0x022: rcDetails{
			"TPM_RC_RETRY",
			"the TPM was not able to start the command",
		},
		0x023: rcDetails{
			"TPM_RC_NV_UNAVAILABLE",
			"the command may require writing of NV and NV is not current accessible",
		},
		0x7F: rcDetails{
			"TPM_RC_NOT_USED",
			"this value is reserved and shall not be returned by the TPM",
		},
	}
)
