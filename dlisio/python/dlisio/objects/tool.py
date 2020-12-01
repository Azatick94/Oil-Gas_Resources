import logging

from .basicobject import BasicObject

class Tool(BasicObject):
    """Tool

    A tool is an ensembly of equiptment that as a whole provide measurements or
    services. The list of equiptment that makes up the tool can be found in
    *tool.parts*. Tools objects also keep a list of all channels that where produced by
    the tool in *tool.channels*.

    Notes
    -----

    The tool object reflects the logical record type TOOL defined in rp66. TOOL
    objects are listed in Appendix A.2 - Logical Record Types, described in
    detail in Chapter 5.8.4 - Static and Frame Data, TOOL objects.

    See also
    --------

    dlisio.Channel : Channel objects.
    dlisio.Parameter : Parameter objects.
    """
    def __init__(self, obj = None, name = None):
        super().__init__(obj, name = name, type = 'TOOL')
        self._description    = None
        self._trademark_name = None
        self._generic_name   = None
        self._status         = None
        self._parts          = []
        self._channels       = []
        self.channels_refs   = []
        self._parameters     = []
        self.parameters_refs = []

    @staticmethod
    def load(obj, name = None):
        self = Tool(obj, name = name)
        for label, value in obj.items():
            if value is None: continue

            if label == "DESCRIPTION"    : self._description    = value[0]
            if label == "TRADEMARK-NAME" : self._trademark_name = value[0]
            if label == "GENERIC-NAME"   : self._generic_name   = value[0]
            if label == "STATUS"         : self._status         = value[0]
            if label == "PARTS"          : self._parts          = value
            if label == "CHANNELS"       : self.channels_refs   = value
            if label == "PARAMETERS"     : self.parameters_refs = value

        self.stripspaces()
        return self

    @property
    def description(self):
        """Description

        Textual description of the Tool.

        Returns
        -------

        description : str
        """
        return self._description

    @property
    def trademark_name(self):
        """Trademark name

        The producer's name for the tool.

        Returns
        -------

        trademark_name : str
        """
        return self._trademark_name

    @property
    def generic_name(self):
        """Generic name

        The name generally used by the industry to describe such a tool.

        Returns
        -------

        generic_name : str
        """
        return self._generic_name

    @property
    def status(self):
        """Status

        Specifies whether the tool is enabled to provide information to the
        acquisition system or not.

        Returns
        -------

        status : int
            Returns 0 if tool is disabled, 1 if it is enabled. Other values
            are undefined.
        """
        return self._status

    @property
    def parts(self):
        """Parts

        A list of Equiptment objects which represent the parts that the tool is
        made up by.

        Returns
        -------

        parts : list of dlisio.core.obname
            each element is a reference to an equiptment object
        """
        return self._parts

    @property
    def channels(self):
        """Channels

        Channels that are produced by this tool.

        Returns
        -------

        channels : list of dlisio.Channel
        """
        return self._channels

    @property
    def parameters(self):
        """Parameters

        Parameter that directly affect or reflect the operation of the tool.

        Returns
        -------

        parameters : list of dlisio.Parameter
        """
        return self._parameters

    def link(self, objects, sets):
        channels = sets['CHANNEL']
        chans = []
        for ref in self.channels_refs:
            fp = ref.fingerprint('CHANNEL')
            try:
                chans.append(channels[fp])
            except KeyError:
                msg = 'missing channel {} referenced from tool {}'
                logging.warning(msg.format(ref, self.name))

        parameters = sets['PARAMETER']
        params = []
        for ref in self.parameters_refs:
            fp = ref.fingerprint('PARAMETER')
            try:
                params.append(parameters[fp])
            except KeyError:
                msg = 'missing parameter {} referenced from tool {}'
                logging.warning(msg.format(ref, self.name))

        self._channels = chans
        self._parameters = params
