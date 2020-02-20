import call from '../../component-loader';

let patientId;
let interval;

export default {
  render: async (patient) => {
    patientId = patient.id;

    if ( interval ) window.clearInterval(interval);
    interval = window.setInterval(() => renderLogs(patientId), 3000);

    await renderLogs(patientId);
  }
}

function renderLogs(patientId) {
  const element = document.getElementById('patient-logs');
  return call(`/api/accessLog/byPatientId/${patientId}`, element)
  .then(logs => {
    element.innerHTML = template(logs);
  });
}

const template = (logs) => `
  <table class="table table-borderless table-bordered table-hover">
    <thead class="thead-dark">
      <tr>
        <th>Timestamp</th>
        <th>Organisation</th>
        <th>Person</th>
      </tr>
    </thead>
    <tbody>
      ${logs.length > 0 ? logs.map(log => `
        <tr>
          <td>${new Date(log.timestamp).toLocaleString('nl-NL')}</td>
          <td>${log.actor.name}</td>
          <td>AGB code ${log.user['irma-demo.nuts.agb.agbcode']}</td>
        </tr>
      `).join('') : '<tr><td colspan="3" style="text-align: center;"><em>None</em></td></tr>'}
    </tbody>
  </table>
`;
