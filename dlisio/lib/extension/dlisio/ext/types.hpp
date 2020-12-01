#ifndef DLISIO_EXT_TYPES_HPP
#define DLISIO_EXT_TYPES_HPP

#include <complex>
#include <cstdint>
#include <exception>
#include <tuple>
#include <type_traits>
#include <utility>
#include <vector>

#include <mpark/variant.hpp>

#include <dlisio/types.h>

#include "strong-typedef.hpp"

namespace dl {

struct not_implemented : public std::logic_error {
    explicit not_implemented( const std::string& msg ) :
        logic_error( "Not implemented yet: " + msg )
    {}
};

struct not_found : public std::runtime_error {
    explicit not_found( const std::string& msg )
        : runtime_error( msg )
    {}
};

enum class representation_code : std::uint8_t {
    fshort = DLIS_FSHORT,
    fsingl = DLIS_FSINGL,
    fsing1 = DLIS_FSING1,
    fsing2 = DLIS_FSING2,
    isingl = DLIS_ISINGL,
    vsingl = DLIS_VSINGL,
    fdoubl = DLIS_FDOUBL,
    fdoub1 = DLIS_FDOUB1,
    fdoub2 = DLIS_FDOUB2,
    csingl = DLIS_CSINGL,
    cdoubl = DLIS_CDOUBL,
    sshort = DLIS_SSHORT,
    snorm  = DLIS_SNORM,
    slong  = DLIS_SLONG,
    ushort = DLIS_USHORT,
    unorm  = DLIS_UNORM,
    ulong  = DLIS_ULONG,
    uvari  = DLIS_UVARI,
    ident  = DLIS_IDENT,
    ascii  = DLIS_ASCII,
    dtime  = DLIS_DTIME,
    origin = DLIS_ORIGIN,
    obname = DLIS_OBNAME,
    objref = DLIS_OBJREF,
    attref = DLIS_ATTREF,
    status = DLIS_STATUS,
    units  = DLIS_UNITS,
};

/*
 * It's _very_ often necessary to access the raw underlying type of the strong
 * type aliases for comparisons, literals, or conversions. dl::decay inspects
 * the argument type and essentially static casts it, regardless of which dl
 * type comes in - it's an automation of static_cast<const value_type&>(x)
 * which otherwise would be repeated a million times
 *
 * The stong typedef's using from detail::strong_typedef has a more specific
 * overload available - the templated version is for completeness.
 */
template < typename T >
T& decay( T& x ) noexcept (true) {
    return x;
}

#define DLIS_REGISTER_TYPEALIAS(name, type) \
    struct name : detail::strong_typedef< name, type > { \
        name() = default; \
        name( const name& ) = default; \
        name( name&& )      = default; \
        name& operator = ( const name& ) = default; \
        name& operator = ( name&& )      = default; \
        using detail::strong_typedef< name, type >::strong_typedef; \
        using detail::strong_typedef< name, type >::operator =; \
        using detail::strong_typedef< name, type >::operator ==; \
        using detail::strong_typedef< name, type >::operator !=; \
    }; \
    inline const type& decay(const name& x) noexcept (true) { \
        return static_cast< const type& >(x); \
    } \
    inline type& decay(name& x) noexcept (true) { \
        return static_cast< type& >(x); \
    }

DLIS_REGISTER_TYPEALIAS(fshort, float)
DLIS_REGISTER_TYPEALIAS(isingl, float)
DLIS_REGISTER_TYPEALIAS(vsingl, float)
DLIS_REGISTER_TYPEALIAS(uvari,  std::int32_t)
DLIS_REGISTER_TYPEALIAS(origin, std::int32_t)
DLIS_REGISTER_TYPEALIAS(ident,  std::string)
DLIS_REGISTER_TYPEALIAS(ascii,  std::string)
DLIS_REGISTER_TYPEALIAS(units,  std::string)
DLIS_REGISTER_TYPEALIAS(status, std::uint8_t)

#undef DLIS_REGISTER_TYPEALIAS

template< typename T, int > struct validated;
template< typename T >
struct validated< T, 2 > {
    T V, A;

    bool operator == (const validated& o) const noexcept (true) {
        return this->V == o.V && this->A == o.A;
    }

    bool operator != (const validated& o) const noexcept (true) {
        return !(*this == o);
    }
};
template< typename T > struct validated< T, 3 > {
    T V, A, B;

    bool operator == (const validated& o) const noexcept (true) {
        return this->V == o.V && this->A == o.A && this->B == o.B;
    }

    bool operator != (const validated& o) const noexcept (true) {
        return !(*this == o);
    }
};

using fsing1 = validated< float, 2 >;
using fsing2 = validated< float, 3 >;
using fdoub1 = validated< double, 2 >;
using fdoub2 = validated< double, 3 >;

using ushort = std::uint8_t;
using unorm  = std::uint16_t;
using ulong  = std::uint32_t;

using sshort = std::int8_t;
using snorm  = std::int16_t;
using slong  = std::int32_t;

using fsingl = float;
using fdoubl = double;

using csingl = std::complex< fsingl >;
using cdoubl = std::complex< fdoubl >;

struct dtime {
    int Y, TZ, M, D, H, MN, S, MS;

    bool operator == (const dtime& o) const noexcept (true) {
        return this->Y  == o.Y
            && this->TZ == o.TZ
            && this->M  == o.M
            && this->D  == o.D
            && this->H  == o.H
            && this->MN == o.MN
            && this->S  == o.S
            && this->MS == o.MS
        ;
    }

    bool operator != (const dtime& o) const noexcept (true) {
        return !(*this == o);
    }
};

struct obname {
    dl::origin origin;
    dl::ushort copy;
    dl::ident  id;

    bool operator == ( const obname& rhs ) const noexcept (true) {
        return this->origin == rhs.origin
            && this->copy == rhs.copy
            && this->id == rhs.id;
    }

    bool operator != (const obname& o) const noexcept (true) {
        return !(*this == o);
    }

    std::string fingerprint(const std::string& type) const noexcept (false);
};

struct objref {
    dl::ident  type;
    dl::obname name;

    bool operator == ( const objref& rhs ) const noexcept( true ) {
        return this->type == rhs.type
            && this->name == rhs.name;
    }

    bool operator != (const objref& o) const noexcept (true) {
        return !(*this == o);
    }

    std::string fingerprint() const noexcept (false);
};

struct attref {
    dl::ident  type;
    dl::obname name;
    dl::ident  label;

    bool operator == ( const attref& rhs ) const noexcept( true ) {
        return this->type == rhs.type
            && this->name == rhs.name
            && this->label== rhs.label;
    }

    bool operator != (const attref& o) const noexcept (true) {
        return !(*this == o);
    }
};

/*
 * Register useful compile time information on the types for other template
 * functions to hook into
 */

template < typename T > struct typeinfo;
template <> struct typeinfo< dl::fshort > {
    static const representation_code reprc = dl::representation_code::fshort;
    constexpr static const char* name = "fshort";
};
template <> struct typeinfo< dl::fsingl > {
    static const representation_code reprc = dl::representation_code::fsingl;
    constexpr static const char* name = "fsingl";
};
template <> struct typeinfo< dl::fsing1 > {
    static const representation_code reprc = dl::representation_code::fsing1;
    constexpr static const char* name = "fsing1";
};
template <> struct typeinfo< dl::fsing2 > {
    static const representation_code reprc = dl::representation_code::fsing2;
    constexpr static const char* name = "fsing2";
};
template <> struct typeinfo< dl::isingl > {
    static const representation_code reprc = dl::representation_code::isingl;
    constexpr static const char* name = "isingl";
};
template <> struct typeinfo< dl::vsingl > {
    static const representation_code reprc = dl::representation_code::vsingl;
    constexpr static const char* name = "vsingl";
};
template <> struct typeinfo< dl::fdoubl > {
    static const representation_code reprc = dl::representation_code::fdoubl;
    constexpr static const char* name = "fdoubl";
};
template <> struct typeinfo< dl::fdoub1 > {
    static const representation_code reprc = dl::representation_code::fdoub1;
    constexpr static const char* name = "fdoub1";
};
template <> struct typeinfo< dl::fdoub2 > {
    static const representation_code reprc = dl::representation_code::fdoub2;
    constexpr static const char* name = "fdoub2";
};
template <> struct typeinfo< dl::csingl > {
    static const representation_code reprc = dl::representation_code::csingl;
    constexpr static const char* name = "csingl";
};
template <> struct typeinfo< dl::cdoubl > {
    static const representation_code reprc = dl::representation_code::cdoubl;
    constexpr static const char* name = "cdoubl";
};
template <> struct typeinfo< dl::sshort > {
    static const representation_code reprc = dl::representation_code::sshort;
    constexpr static const char* name = "sshort";
};
template <> struct typeinfo< dl::snorm > {
    static const representation_code reprc = dl::representation_code::snorm;
    constexpr static const char* name = "snorm";
};
template <> struct typeinfo< dl::slong > {
    static const representation_code reprc = dl::representation_code::slong;
    constexpr static const char* name = "slong";
};
template <> struct typeinfo< dl::ushort > {
    static const representation_code reprc = dl::representation_code::ushort;
    constexpr static const char* name = "ushort";
};
template <> struct typeinfo< dl::unorm > {
    static const representation_code reprc = dl::representation_code::unorm;
    constexpr static const char* name = "unorm";
};
template <> struct typeinfo< dl::ulong > {
    static const representation_code reprc = dl::representation_code::ulong;
    constexpr static const char* name = "ulong";
};
template <> struct typeinfo< dl::uvari > {
    static const representation_code reprc = dl::representation_code::uvari;
    constexpr static const char* name = "uvari";
};
template <> struct typeinfo< dl::ident > {
    static const representation_code reprc = dl::representation_code::ident;
    constexpr static const char* name = "ident";
};
template <> struct typeinfo< dl::ascii > {
    static const representation_code reprc = dl::representation_code::ascii;
    constexpr static const char* name = "ascii";
};
template <> struct typeinfo< dl::dtime > {
    static const representation_code reprc = dl::representation_code::dtime;
    constexpr static const char* name = "dtime";
};
template <> struct typeinfo< dl::origin > {
    static const representation_code reprc = dl::representation_code::origin;
    constexpr static const char* name = "origin";
};
template <> struct typeinfo< dl::obname > {
    static const representation_code reprc = dl::representation_code::obname;
    constexpr static const char* name = "obname";
};
template <> struct typeinfo< dl::objref > {
    static const representation_code reprc = dl::representation_code::objref;
    constexpr static const char* name = "objref";
};
template <> struct typeinfo< dl::attref > {
    static const representation_code reprc = dl::representation_code::attref;
    constexpr static const char* name = "attref";
};
template <> struct typeinfo< dl::status > {
    static const representation_code reprc = dl::representation_code::status;
    constexpr static const char* name = "status";
};
template <> struct typeinfo< dl::units > {
    static const representation_code reprc = dl::representation_code::units;
    constexpr static const char* name = "units";
};

/*
 * Parsing and parsing input
 *
 * the strategy is to first parse the EFLR template and build a parsing guide,
 * expressed as the object_template. Later, this template instantiates the
 * default object in the set, and edits the fields as it goes along. The value
 * field can be zero or more values, so it's neatly stored as a vector, but the
 * *type* is indeterminate until the representation code is understood.
 *
 * The variant is a perfect fit for this.
 *
 * A variant-of-vector seems a better fit than vector-of-variants, both because
 * the max-size-overhead isn't so bad (all vectors are the same size), but the
 * type-resolution only has to be done once, and the unstructuring of the
 * vector can be contained inside the visitor.
 */
using value_vector = mpark::variant<
    mpark::monostate,
    std::vector< fshort >,
    std::vector< fsingl >,
    std::vector< fsing1 >,
    std::vector< fsing2 >,
    std::vector< isingl >,
    std::vector< vsingl >,
    std::vector< fdoubl >,
    std::vector< fdoub1 >,
    std::vector< fdoub2 >,
    std::vector< csingl >,
    std::vector< cdoubl >,
    std::vector< sshort >,
    std::vector< snorm  >,
    std::vector< slong  >,
    std::vector< ushort >,
    std::vector< unorm  >,
    std::vector< ulong  >,
    std::vector< uvari  >,
    std::vector< ident  >,
    std::vector< ascii  >,
    std::vector< dtime  >,
    std::vector< origin >,
    std::vector< obname >,
    std::vector< objref >,
    std::vector< attref >,
    std::vector< status >,
    std::vector< units  >
>;

/*
 * The structure of an attribute as described in 3.2.2.1
 */
struct object_attribute {
    dl::ident           label = {};
    dl::uvari           count = dl::uvari{ 1 };
    representation_code reprc = representation_code::ident;
    dl::units           units = {};
    dl::value_vector    value = {};
    bool invariant            = false;

    bool operator == (const object_attribute& ) const noexcept (true);
};

/*
 * The Object Set Template (3.2.1 EFLR: General layout) is just an ordered set
 * of attributes, so just alias a vector
 */
using object_template = std::vector< object_attribute >;

/*
 * Parsing output and semantic objects
 *
 * C++ representation of the set of logical record types (listed in Appendix A
 * - Logical Record Types, described in Chapter 5 - Static and Frame Data)
 *
 * They all derive from basic_object, but that's just a low-syntactical
 * overhead way of adding the object-name field, which is present in every
 * object. In fact, this object-name preceeds the attributes and is introduced
 * by the component descriptor (3.2.2.1 Component Descriptor figure 3-4). It
 * carries no other semantic or operational significance and should be
 * considered an implementation detail.
 *
 * While not very clear on the matter, in practice every attribute of even
 * well-specified object types (such as CHANNEL) can be absent. Because of
 * this, every attribute is an Optional. The lack of a value in the optional
 * means either that the attribute is explicitly marked absent (see
 * DLIS_ROLE_ABSATR), or not present at all in the object template. It is
 * impossible to distinguish these cases without consulting the template
 * itself.
 *
 * The member variables are designed and enriched with the intention that
 * member variables are set with the object_attribute::into functions. Other
 * code expects that all objects have the two methods set and remove:
 *
 * void set( const object_attribute& );
 * void remove( const object_attribute& );
 *
 * These should map object attribute label to the right member variable, and
 * set/unset respectively - remove is called when encoutering ABSATR, set
 * otherwise.
 */

/*
 * All objects have an object name (3.2.2.1 Component Descriptor figure 3-4)
 */
struct basic_object {
    void set( const object_attribute& )    noexcept (false);
    void remove( const object_attribute& ) noexcept (false);

    std::size_t len() const noexcept (true);
    const dl::object_attribute&
    at( const std::string& ) const noexcept (false);

    bool operator == (const basic_object&) const noexcept (true);
    bool operator != (const basic_object&) const noexcept (true);

    dl::obname object_name;
    std::vector< object_attribute > attributes;
};

/*
 * The object set, after parsing, is an (unordered?) collection of objects. In
 * parsing, more information is added through creating custom types, but the
 * fundamental restriction is one type per set.
 *
 * The variant-of-vectors is wells suited for this
 */
using object_vector = std::vector< basic_object >;

struct object_set {
    int role; // TODO: enum class?
    dl::ident type;
    dl::ident name;
    dl::object_template tmpl;
    dl::object_vector objects;
};

const char* parse_template( const char* begin,
                            const char* end,
                            object_template& ) noexcept (false);


object_set parse_objects( const char*, const char* ) noexcept (false);

}

#endif //DLISIO_EXT_TYPES_HPP
