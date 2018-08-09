import Button from '@material-ui/core/Button';
import Checkbox from '@material-ui/core/Checkbox';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogTitle from '@material-ui/core/DialogTitle';
import ExpansionPanel from '@material-ui/core/ExpansionPanel';
import ExpansionPanelDetails from '@material-ui/core/ExpansionPanelDetails';
import ExpansionPanelSummary from '@material-ui/core/ExpansionPanelSummary';
import Grid from '@material-ui/core/Grid';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import Paper from '@material-ui/core/Paper';
import { withStyles } from '@material-ui/core/styles';
import Typography from '@material-ui/core/Typography';
import AddIcon from '@material-ui/icons/Add';
import DeleteIcon from '@material-ui/icons/Delete';
import SubmitIcon from '@material-ui/icons/Done';
import CloudIcon from '@material-ui/icons/Cloud';
import ExpandMoreIcon from '@material-ui/icons/ExpandMore';
import 'brace/mode/javascript';
import 'brace/theme/github';
import React, { Component } from 'react';
import AceEditor from 'react-ace';
import Form from "react-jsonschema-form";
import uuidv4 from "uuid/v4";


const styles = theme => ({
  root: {
    flexGrow: 1,
  },
  paper: {
    padding: theme.spacing.unit * 2,
    textAlign: 'center',
    color: theme.palette.text.secondary,
    cursor: 'pointer'
  },
  heading: {
    fontSize: '12px',
    fontWeight: theme.typography.fontWeightRegular,
  },
  button: {
    margin: theme.spacing.unit,
  },
  leftIcon: {
    marginRight: theme.spacing.unit,
  },
  rightIcon: {
    marginLeft: theme.spacing.unit,
  },
  iconSmall: {
    fontSize: 20,
  },
});


const eventSchema = {
  type: "object",
  eventID: "",
  delay: "",
  source: "",
  eventTime:"",
  data: "",
  eventType: "",
  required: ["eventID", "delay", "data", "source", "eventType","eventTime"],
  properties: {
    delay: { type: "string", title: "Submit Delay(ms)", default: "1000" },
    eventType: { type: "string", title: "Event Type", default: "acme.prod.site247.search_down" },
    source: { type: "string", title: "Event Source", default: "site247" },
    data: { type: "string", title: "Event Data", default: '{"appinfoA":"abc","appinfoB": 123,"appinfoC": true}' },
    eventID: { type: "string", title: "Event ID", default: uuidv4() },
    eventTime: { type: "string", title: "Event Time", default: "2018-04-05T17:31:00Z" }
  }
};

const eventSchemaCreate = {
  type: "object",
  delay: "",
  source: "",
  eventTime:"",
  data: "",
  eventType: "",
  required: ["eventID", "delay", "data", "source", "eventType","eventTime"],
  properties: {
    delay: { type: "string", title: "Submit Delay(ms)", default: "1000" },
    eventType: { type: "string", title: "Event Type", default: "acme.prod.site247.search_down" },
    source: { type: "string", title: "Event Source", default: "site247" },
    data: { type: "string", title: "Event Data", default: '{"appinfoA":"abc","appinfoB": 123,"appinfoC": true}' },
    eventTime: { type: "string", title: "Event Time", default: "2018-04-05T17:31:00Z" }
  }
};

const uiSchemaEvent = {
  data: {
    "ui:widget": "textarea"
  },
  delay: {
    "ui:help": "Hint: Delay in ms for submission delay and it's not part of cloudevents.io spec"
  },
  "ui:widget": "string",
  "ui:help": "Hint: Create event with delay in ms associated with it"
};

const EventCard = (props) => {
  const { classes, event, handleChangeEvent, handlePannelExpansion,handleSubmitEvent } = props;
  return (
    <ExpansionPanel onChange={(e, flag) => handlePannelExpansion(e, flag, event.eventID)}>
      <ExpansionPanelSummary expandIcon={<ExpandMoreIcon />}>
        <Typography className={classes.heading}>{event && event.eventType}</Typography>
      </ExpansionPanelSummary>
      <ExpansionPanelDetails>
        <div style={{ width: '100%', fontSize: '12px' }}>
          <Form
            uiSchema={uiSchemaEvent}
            schema={eventSchema}
            formData={event}
            onSubmit={(e) => handleSubmitEvent(e, event.eventID)}
            onChange={(e) => handleChangeEvent(e, event.eventID)}
            onError={log("errors")}>
          </Form>
        </div>
      </ExpansionPanelDetails>
    </ExpansionPanel>
  )
}

const log = (type) => console.log.bind(console, type);

class Sandbox extends Component {
  constructor(props, context) {
    super(props, context);
    let eventList = localStorage.getItem("events");
    try{
      eventList = JSON.parse(eventList);
    }
    catch(e){
      console.log("Unable to parse data from local storage",e)
    }
    this.state = {
      events: eventList?eventList:[],
      newEvent: "",
      eventData: "",
      ruleDialogOpen: false,
      eventsChecked: [],
      selectedEvent: '',
      expansionFlag: false,
      result:[],
      resultFlag:false,
      cloudDialogOpen:false,
      cloudEvent:"",
      modalFlag:false,
      alertType:'warning',
      response:""
    }
  }

  handleEventDialogSave = () => {
    let { events, newEvent } = this.state;
    let flag = false;
    let condition = false;
    if(newEvent === ""){
      this.setState({
        response:"Event cannot be created with all default values",
        modalFlag:true,
        alertType:'warning'
      });
      return
    }
    for(let x in Object.keys(newEvent)){
      if(newEvent[x] === ""){
        this.setState({
          response: x + " cannot be empty",
          modalFlag:true,
          alertType:'warning'
        });
        condition = true;
        return;
      }
    }
    if(condition){
      return;
    }
    let data;
    try {
      data = JSON.parse(newEvent.data);
    }
    catch(e) {
      this.setState({
        response:"Incorrect JSON data",
        modalFlag:true,
        alertType:'warning'
      });
      return
    }
    events.forEach((item) => {
      if (item.eventID === newEvent.eventID) {
        flag = true;
      }
    })
    if(flag){
      this.setState({
        response:"Event name already exists",
        modalFlag:true,
        alertType:'warning'
      });
      return
    }
    newEvent["eventID"] = uuidv4();
    events.push(newEvent);
    localStorage.setItem("events",JSON.stringify(events));
    this.setState({
      events: events,
      newEvent: {},
      ruleDialogOpen: false
    })
  }

  handleEventsCheckToggle = (value) => {
    const { eventsChecked } = this.state;
    const currentIndex = eventsChecked.indexOf(value);
    const newChecked = [...eventsChecked];

    if (currentIndex === -1) {
      newChecked.push(value);
    } else {
      newChecked.splice(currentIndex, 1);
    }

    this.setState({
      eventsChecked: newChecked,
    });
  }

  handleEventChange = (value) => {
    const { selectedEvent, events } = this.state;
    let obj;
    let eventList = events.map((item, i) => {
      if (item.eventID === selectedEvent.eventID) {
        item.data = value;
        obj = item;
        return obj;
      } else {
        return item;
      }
    });
    localStorage.setItem("events",JSON.stringify(eventList))
    this.setState({
      selectedEvent: obj,
      events: eventList
    })
  }

  handleEventsExecution = () => {
    let self = this;
    const { eventsChecked,events,result } = this.state;
    let eventsList = [];
    if (eventsChecked.length < 1) {
      this.setState({
        response:"No event selected for execution",
        modalFlag:true,
        alertType:'warning'
      });
      return
    }
    events.forEach((item) => {
      if(eventsChecked.indexOf(item.eventID) !== -1){
        let data;
        try {
          data = JSON.parse(item.data);
        }
        catch(e) {
          this.setState({
            response:"Incorrect JSON data for event ID " + item.eventID,
            modalFlag:true,
            alertType:'warning'
          });
          return
        }
        eventsList.push(item);
      }
    })
    this.setState({resultFlag:true})
    eventsList.forEach((event)=>{
      let json = self.cloudEventFormat(event);
      let time = parseInt(event.delay,10);
      setTimeout(function(jsonData) {
        // API calls for updating events
        fetch('/event', {
          method: "POST",
          body: JSON.stringify(jsonData)
        })
        .then(function (response) {
          if (response.ok) {
            let resultSet = result;
            resultSet.push({"id":jsonData.eventID,"output":" Successfully called API with delay of " + time + " ms"});
            self.setState({
              result:resultSet
            })
            console.log("Updated successfully");
          } else {
            throw new Error('Something went wrong. Unable to push event');
          }
        })
        .catch((error) => {
          let resultSet = result;
          resultSet.push({"id":jsonData.eventID,"output":error});
          self.setState({
            result:resultSet
          })
        });
      }, time, json);
    })
  }

  handleEventExpansion = (event, flag, id) => {
    const { events } = this.state;
    let obj;
    events.forEach((item) => {
      if (item.eventID === id) {
        obj = item;
      }
    })
    if (!obj) {
      this.setState({
        response:"No event found",
        modalFlag:true,
        alertType:'warning'
      });
      return
    }
    this.setState({
      selectedEvent: obj,
      expansionFlag: flag
    })
  }

  cloudEventFormat = (obj) =>{
    let data;
    let flag = false;
    if(typeof obj === 'string'){
      try {
        obj = JSON.parse(obj);
      }
      catch(e) {
        this.setState({
          response:"Incorrect JSON data",
          modalFlag:true,
          alertType:'warning'
        });
        flag = true;
        return
      }
      data = obj.data;
    }
    else{
      data = JSON.parse(obj.data);
    }
    if(flag){
      return;
    }
    let eventDummy = {
      "cloudEventsVersion": "0.1",
      "eventType": obj.eventType,
      "source": obj.source,
      "eventID": obj.eventID,
      "eventTime": obj.eventTime,
      "extensions": {
        "comExampleExtension": "value"
      },
      "contentType": "application/json",
      "data": data
    }
    return eventDummy;
  }

  handleOldEventChange = (event, id) => {
    let obj = event.formData;
    this.setState({selectedEvent:obj})
  }

  handleEventChangeSubmit = (event, id) => {
    const {events} = this.state
    let obj = event.formData;
    let newList = events.map((item, i) => {
      if (item.eventID === id) {
        item = obj;
      }
      return item
    })
    this.setState({
      response:"Event Updated",
      modalFlag:true,
      alertType:'confirmation'
    });
    localStorage.setItem("events",JSON.stringify(newList));
    this.setState({ events: newList, selectedEvent: obj })
  }

  handleEventDelete = () => {
    const {eventsChecked,events} = this.state
    if(eventsChecked.length < 1){
      this.setState({
        response:"No event selected for deletion",
        modalFlag:true,
        alertType:'warning'
      });
      return
    }
    let updatedEvents = [];
    events.forEach((item) => {
      if(eventsChecked.indexOf(item.eventID) !== -1){
        console.log('Item',item);
      }
      else{
        updatedEvents.push(item);
      }
    })
    localStorage.setItem("events",JSON.stringify(updatedEvents))
    this.setState({events:updatedEvents,eventsChecked:[],selectedEvent:""});
  }

  handleScriptChange = (text,event) => {
    this.setState({selectedEvent:text})
  }

  handleCloudDialogSave = () => {
    const { cloudEvent,events } = this.state;
    let newevent;
    try {
      newevent = JSON.parse(cloudEvent);
    } catch (error) {
      this.setState({
        response:"Cannot parse the JSON",
        modalFlag:true,
        alertType:'warning'
      });
      return
    }
    newevent["delay"] = 1000;
    newevent["EventID"] = uuidv4();

    let eventattr = ["cloudEventsVersion","eventType","source","eventID","eventTime","extensions","contentType","data"]
    let eventFlag = false;
    let eventKeys = Object.keys(cloudEvent);
    for (let attr in eventattr){
      if(eventKeys.indexOf(attr) === -1){
        eventFlag = true
      }
    }

    newevent["data"] = JSON.stringify(newevent["data"]);

    if(eventFlag){
      this.setState({
        response:"Not a valid cloud event",
        modalFlag:true,
        alertType:'warning'
      });
      return
    }
    events.push(newevent);
    localStorage.setItem("events",JSON.stringify(events))
    this.setState({
      events:events,
      cloudEvent:"",
      cloudDialogOpen:false
    })
  }

  render() {
    const { classes } = this.props;
    const { selectedEvent, events, cloudEvent, response } = this.state;
    let self = this;
    return (
      <div>
        <Dialog
          fullScreen={false}
          fullWidth={true}
          open={this.state.ruleDialogOpen}
          onClose={() => this.setState({ ruleDialogOpen: false })}
          aria-labelledby="responsive-dialog-title"
        >
          <DialogTitle id="responsive-dialog-title">Create New Event</DialogTitle>
          <DialogContent>
            <Form
              uiSchema={uiSchemaEvent}
              schema={eventSchemaCreate}
              formData={this.state.newEvent}
              onChange={(event) => this.setState({ newEvent: event.formData })}
              onError={(error) => console.log("errors", error)} >
              <button type="submit" className="hidden">Submit</button>
            </Form>
          </DialogContent>
          <DialogActions>
            <Button onClick={() => this.setState({ ruleDialogOpen: false })}
              style={{fontSize:'12px'}} 
              color="primary">
              Cancel
              </Button>
            <Button onClick={this.handleEventDialogSave}
              style={{fontSize:'12px'}} 
              color="primary" autoFocus>
              Save
            </Button>
          </DialogActions>
        </Dialog>
        <Dialog
          fullScreen={false}
          fullWidth={true}
          open={this.state.resultFlag}
          onClose={() => this.setState({ resultFlag: false })}
          aria-labelledby="responsive-dialog-title"
        >
          <DialogTitle id="responsive-dialog-title">Event Execution Results</DialogTitle>
          <DialogContent>
            {
              this.state.result.map((item) => (
                <div key={item.id}>
                  <span style={{fontWeight:'700',marginRight:'5px'}}>EventID</span> {item.id}
                  <span style={{fontWeight:'700',marginRight:'5px',marginLeft:'5px'}}>Result</span> {item.output}
                </div>
              ))
            }
          </DialogContent>
          <DialogActions>
            <Button onClick={() => this.setState({ resultFlag: false,result:[] })}
              style={{fontSize:'12px'}} 
              color="primary">
              Close
            </Button>
          </DialogActions>
        </Dialog>
        <Dialog
          fullScreen={false}
          fullWidth={true}
          open={this.state.cloudDialogOpen}
          onClose={() => this.setState({ cloudDialogOpen: false })}
          aria-labelledby="responsive-dialog-title"
        >
          <DialogTitle id="responsive-dialog-title">Event Cloud Event</DialogTitle>
          <DialogContent>
            {/* Create New Cloud Event */}
            <AceEditor
              mode="javascript"
              theme="github"
              name="Cloud Event"
              fontSize={14}
              showPrintMargin={true}
              onChange={(text, event) => this.setState({cloudEvent:text})}
              showGutter={true}
              highlightActiveLine={true}
              value={cloudEvent}
              setOptions={{
                enableBasicAutocompletion: false,
                enableLiveAutocompletion: false,
                enableSnippets: true,
                showLineNumbers: true,
                tabSize: 2,
              }} />
          </DialogContent>
          <DialogActions>
            <Button onClick={() => this.setState({ cloudDialogOpen: false })}
              style={{fontSize:'12px'}} 
              color="primary">
              Close
            </Button>
            <Button onClick={this.handleCloudDialogSave}
              style={{fontSize:'12px'}} 
              color="primary" autoFocus>
              Save
            </Button>
          </DialogActions>
        </Dialog>
        <Grid container >
          <Button onClick={() => this.setState({ ruleDialogOpen: true })}
            variant="contained"
            color="primary"
            className={classes.button}>
            Add
            <AddIcon className={classes.rightIcon} />
          </Button>
          <Button onClick={this.handleEventDelete}
            variant="contained"
            color="secondary"
            className={classes.button}>
            Delete
            <DeleteIcon className={classes.rightIcon} />
          </Button>
          <Button onClick={() => this.setState({ cloudDialogOpen: true })}
            variant="contained"
            color="primary"
            className={classes.button}>
            Cloud Event
            <CloudIcon className={classes.rightIcon} />
          </Button>
          <Button onClick={this.handleEventsExecution}
            style={{ marginLeft: 'auto' }}
            variant="contained"
            color="primary"
            className={classes.button}>
            Execute
            <SubmitIcon className={classes.rightIcon} />
          </Button>
        </Grid>
        <Grid container spacing={24}>
          <Grid item xs={4}>
            <List>
              {(events.length>0)?
                events.map((event, index) => (
                <Grid key={index} container>
                  <Grid item xs={2}>
                    <ListItem
                      key={index}
                      role={undefined}
                      onClick={() => self.handleEventsCheckToggle(event.eventID)}
                      className={classes.listItem}
                    >
                      <Checkbox
                        checked={self.state.eventsChecked.indexOf(event.eventID) !== -1}
                        tabIndex={-1}
                        disableRipple
                      />
                    </ListItem>
                  </Grid>
                  <Grid item xs={10} >
                    <EventCard key={index}
                      handlePannelExpansion={self.handleEventExpansion}
                      handleSubmitEvent={(event, id) => self.handleEventChangeSubmit(event,id)}
                      handleChangeEvent={(event, id) => self.handleOldEventChange(event, id)}
                      classes={classes} event={(event.eventID === selectedEvent.eventID)?selectedEvent:event} />
                  </Grid>
                </Grid>
              )):false}
            </List>
          </Grid>
          <Grid item xs={8}>
            {(this.state.selectedEvent !== '' && this.state.expansionFlag) ?
              <Paper className={classes.paper}>
                <AceEditor
                  mode="javascript"
                  theme="github"
                  name="blah2"
                  fontSize={14}
                  readOnly={true}
                  showPrintMargin={true}
                  onChange={(text, event) => this.handleScriptChange(text,event)}
                  showGutter={true}
                  highlightActiveLine={true}
                  value={`${JSON.stringify(self.cloudEventFormat(self.state.selectedEvent),null, "\t")}`}
                  setOptions={{
                    enableBasicAutocompletion: false,
                    enableLiveAutocompletion: false,
                    enableSnippets: true,
                    showLineNumbers: true,
                    tabSize: 2,
                  }} />
              </Paper> : false
            }
          </Grid>
        </Grid>
        <Dialog
          fullScreen={false}
          fullWidth={true}
          open={this.state.modalFlag}
          onClose={() => this.setState({ modalFlag: false })}
          aria-labelledby="responsive-dialog-title"
        >
        <DialogTitle id="responsive-dialog-title">Status</DialogTitle>
          <DialogContent>
             {response}
          </DialogContent>
          <DialogActions>
            <Button onClick={() => this.setState({ modalFlag: false })}
              style={{fontSize:'12px'}} 
              color="primary">
              Close
            </Button>
          </DialogActions>
        </Dialog>
      </div>
    )
  }

}

export default withStyles(styles, { withTheme: true })(Sandbox);